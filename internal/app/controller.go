package app

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/middleware"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"ikoyhn/podcast-sponsorblock/internal/services/playlist"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/sponsorblock"
	"ikoyhn/podcast-sponsorblock/internal/services/analytics"
	"ikoyhn/podcast-sponsorblock/internal/services/backup"
	"ikoyhn/podcast-sponsorblock/internal/services/search"
	"ikoyhn/podcast-sponsorblock/internal/api"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func registerRoutes(e *echo.Echo) {
	// Root endpoint - no auth, basic rate limiting
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Health check endpoints - no auth, basic rate limiting
	e.GET("/health", healthCheckHandler, middleware.RateLimitMiddleware(60))
	e.GET("/ready", readinessCheckHandler, middleware.RateLimitMiddleware(60))

	// Metrics endpoint for Prometheus scraping
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Serve static files (CSS, JS, images)
	e.Static("/static", "web/static")

	// Admin panel routes - serve HTML files
	e.GET("/admin", func(c echo.Context) error {
		return c.File("web/admin/dashboard.html")
	})
	e.GET("/admin/", func(c echo.Context) error {
		return c.File("web/admin/dashboard.html")
	})
	e.GET("/admin/podcasts", func(c echo.Context) error {
		return c.File("web/admin/podcasts.html")
	})
	e.GET("/admin/episodes", func(c echo.Context) error {
		return c.File("web/admin/episodes.html")
	})
	e.GET("/admin/settings", func(c echo.Context) error {
		return c.File("web/admin/settings.html")
	})
	e.GET("/admin/analytics", func(c echo.Context) error {
		return c.File("web/admin/analytics.html")
	})

	// Register backup API routes
	api.RegisterBackupRoutes(e)

	// Register recommendation API routes
	api.RegisterRecommendationRoutes(e)

	// Register analytics API routes
	registerAnalyticsRoutes(e)

	// Register transcript API routes
	registerTranscriptRoutes(e)

	// Swagger UI documentation
	e.GET("/docs/*", echoSwagger.WrapHandler)

	// OPML routes
	e.POST("/api/opml/import", api.ImportOPMLHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(5))
	e.GET("/api/opml/export", api.ExportOPMLHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Smart Playlist API routes
	e.POST("/api/playlist/smart", api.CreateSmartPlaylistHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.GET("/api/playlist/smart", api.GetAllSmartPlaylistsHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.GET("/api/playlist/smart/:id", api.GetSmartPlaylistHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.PUT("/api/playlist/smart/:id", api.UpdateSmartPlaylistHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.DELETE("/api/playlist/smart/:id", api.DeleteSmartPlaylistHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Smart Playlist RSS feed route
	e.GET("/rss/smart/:id", api.GetSmartPlaylistRSSHandler, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// User Preferences API routes
	e.GET("/api/preferences", api.GetUserPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.PUT("/api/preferences", api.UpdateUserPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.GET("/api/preferences/feeds", api.GetAllFeedPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.GET("/api/preferences/feed/:feedId", api.GetFeedPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.PUT("/api/preferences/feed/:feedId", api.UpdateFeedPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.DELETE("/api/preferences/feed/:feedId", api.DeleteFeedPreferences, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Content Filter API routes
	e.POST("/api/filters", api.CreateFilter, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.GET("/api/filters", api.GetFilters, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.GET("/api/filters/:id", api.GetFilter, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
	e.PUT("/api/filters/:id", api.UpdateFilter, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.DELETE("/api/filters/:id", api.DeleteFilter, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
	e.PATCH("/api/filters/:id/toggle", api.ToggleFilter, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	e.GET("/channel/:channelId", func(c echo.Context) error {
		rssRequestParams, err := validateQueryParams(c, c.Param("channelId"))
		if err != nil {
			return err
		}
		data := rss.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		rssRequestParams, err := validateQueryParams(c, c.Param("youtubePlaylistId"))
		if err != nil {
			return err
		}
		data := playlist.BuildPlaylistRssFeed(c.Param("youtubePlaylistId"), rssRequestParams, handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Media endpoint - auth + higher rate limiting (50 req/min)
	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
		fileName := c.Param("youtubeVideoId")
		if !common.IsValidParam(fileName) {
			return appErrors.NewInvalidParamError("youtubeVideoId").
				WithDetail("value", fileName)
		}
		if !common.IsValidFilename(fileName) {
			return appErrors.NewFileNotFoundError(fileName)
		}

		// Get format and quality from query parameters
		formatParam := c.QueryParam("format")
		if formatParam == "" {
			formatParam = config.Config.AudioFormat
		}
		audioFormat := downloader.GetAudioFormat(formatParam)

		// Strip extension from fileName if present
		for _, fmt := range []downloader.AudioFormat{downloader.FormatM4A, downloader.FormatMP3, downloader.FormatOpus} {
			fileName = strings.TrimSuffix(fileName, fmt.Extension)
		}

		// Track analytics for this play
		analyticsService := analytics.NewService()
		if err := analyticsService.TrackEpisodePlay(fileName, c.Request()); err != nil {
			logger.Logger.Error().
				Err(err).
				Str("episode_id", fileName).
				Msg("Failed to track episode play")
			// Don't fail the request if analytics tracking fails
		}

		// Try to open file with requested format
		filePath := path.Join(config.Config.AudioDir, fileName+audioFormat.Extension)
		file, err := os.Open(filePath)
		needRedownload, totalTimeSkipped := sponsorblock.DeterminePodcastDownload(fileName)

		if file == nil || err != nil || needRedownload {
			if file != nil {
				file.Close()
			}
			database.UpdateEpisodePlaybackHistory(fileName, totalTimeSkipped)
			fileName, done := downloader.GetYoutubeVideoWithFormat(fileName, audioFormat)
			<-done
			file, err = os.Open(path.Join(config.Config.AudioDir, fileName+audioFormat.Extension))
			if err != nil || file == nil {
				return appErrors.NewFileNotFoundError(fileName+audioFormat.Extension).
					WithDetail("error", "Failed to open file after download")
			}
			defer file.Close()

			rangeHeader := c.Request().Header.Get("Range")
			if rangeHeader != "" {
				http.ServeFile(c.Response().Writer, c.Request(), path.Join(config.Config.AudioDir, fileName+audioFormat.Extension))
				return nil
			}
			return c.Stream(http.StatusOK, audioFormat.MimeType, file)
		}

		defer file.Close()
		database.UpdateEpisodePlaybackHistory(fileName, totalTimeSkipped)
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			http.ServeFile(c.Response().Writer, c.Request(), filePath)
			return nil
		}
		return c.Stream(http.StatusOK, audioFormat.MimeType, file)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(50))

	// Search endpoints - auth + moderate rate limiting (20 req/min)
	registerSearchRoutes(e)

	// Batch operations endpoints - auth + strict rate limiting (5 req/min)
	registerBatchRoutes(e)

	// Webhook configuration endpoints - auth + moderate rate limiting (10 req/min)
	registerWebhookRoutes(e)

	port := os.Getenv("PORT")
	if port == "" {
		port = constants.DefaultPort
	}
	host := os.Getenv("HOST")

	logger.Logger.Info().Str("host", host).Str("port", port).Msg("Starting server")
	logger.Logger.Fatal().Err(e.Start(host + ":" + port)).Msg("Server failed to start")

}
func validateQueryParams(c echo.Context, id string) (*models.RssRequestParams, error) {
	limitVar := c.Request().URL.Query().Get("limit")
	dateVar := c.Request().URL.Query().Get("date")
	if !common.IsValidParam(id) {
		return nil, appErrors.NewInvalidParamError("id").WithDetail("value", id)
	}
	if c.Request().URL.Query().Get("limit") != "" && c.Request().URL.Query().Get("date") != "" {
		return nil, appErrors.NewBadRequestError("Cannot specify both 'limit' and 'date' parameters")
	}

	if limitVar != "" {
		limitInt, err := strconv.Atoi(c.Request().URL.Query().Get("limit"))
		if err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to parse limit parameter")
			return nil, appErrors.NewInvalidParamError("limit").WithDetail("value", limitVar)
		}
		return &models.RssRequestParams{Limit: &limitInt, Date: nil}, nil
	}

	if dateVar != "" {
		parsedDate, err := time.Parse("01-02-2006", dateVar)
		if err != nil {
			logger.Logger.Error().Err(err).Str("date", dateVar).Msg("Failed to parse date string")
			return nil, appErrors.NewInvalidParamError("date").WithDetail("value", dateVar).WithDetail("format", "MM-DD-YYYY")
		}
		return &models.RssRequestParams{Limit: nil, Date: &parsedDate}, nil
	}
	return &models.RssRequestParams{Limit: nil, Date: nil}, nil
}

func setupCron() *cron.Cron {
	c := cron.New()

	// Setup podcast cleanup cron
	cronSchedule := constants.DefaultCron
	if config.Config.Cron != "" {
		cronSchedule = config.Config.Cron
	}
	c.AddFunc(cronSchedule, func() {
		database.DeletePodcastCronJob()
	})

	// Setup backup cron if configured
	if config.Config.BackupCron != "" {
		c.AddFunc(config.Config.BackupCron, func() {
			logger.Logger.Info().Msg("Running scheduled backup")
			if err := backup.ScheduledBackup(config.Config.BackupIncludeAudio); err != nil {
				logger.Logger.Error().Err(err).Msg("Scheduled backup failed")
			}
		})
		logger.Logger.Info().
			Str("schedule", config.Config.BackupCron).
			Bool("include_audio", config.Config.BackupIncludeAudio).
			Msg("Backup cron job scheduled")
	}

	c.Start()
	return c
}

func setupHandlers(e *echo.Echo) {

}

func handler(r *http.Request) string {
	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	host := r.Host
	url := scheme + "://" + host
	return url
}

// registerSearchRoutes registers all search-related routes
func registerSearchRoutes(e *echo.Echo) {
	// Search episodes endpoint
	e.GET("/search/episodes", func(c echo.Context) error {
		var req models.SearchEpisodeRequest
		if err := c.Bind(&req); err != nil {
			return appErrors.NewBadRequestError("Invalid query parameters").
				WithDetail("error", err.Error())
		}

		// Validate and parse parameters
		searchReq := search.EpisodeSearchRequest{
			Query:  req.Query,
			Limit:  req.Limit,
			Offset: req.Offset,
		}

		// Parse podcast_id filter
		if req.PodcastId != "" {
			if !common.IsValidParam(req.PodcastId) {
				return appErrors.NewInvalidParamError("podcast_id").
					WithDetail("value", req.PodcastId)
			}
			searchReq.PodcastId = &req.PodcastId
		}

		// Parse type filter
		if req.Type != "" {
			validTypes := map[string]bool{"CHANNEL": true, "PLAYLIST": true}
			if !validTypes[req.Type] {
				return appErrors.NewInvalidParamError("type").
					WithDetail("value", req.Type).
					WithDetail("allowed", "CHANNEL, PLAYLIST")
			}
			searchReq.Type = &req.Type
		}

		// Parse date filters
		if req.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", req.StartDate)
			if err != nil {
				return appErrors.NewInvalidParamError("start_date").
					WithDetail("value", req.StartDate).
					WithDetail("format", "YYYY-MM-DD")
			}
			searchReq.StartDate = &startDate
		}

		if req.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", req.EndDate)
			if err != nil {
				return appErrors.NewInvalidParamError("end_date").
					WithDetail("value", req.EndDate).
					WithDetail("format", "YYYY-MM-DD")
			}
			searchReq.EndDate = &endDate
		}

		// Parse duration filters
		if req.MinDuration != "" {
			minDuration, err := time.ParseDuration(req.MinDuration)
			if err != nil {
				return appErrors.NewInvalidParamError("min_duration").
					WithDetail("value", req.MinDuration).
					WithDetail("format", "duration string (e.g., 30m, 1h)")
			}
			searchReq.MinDuration = &minDuration
		}

		if req.MaxDuration != "" {
			maxDuration, err := time.ParseDuration(req.MaxDuration)
			if err != nil {
				return appErrors.NewInvalidParamError("max_duration").
					WithDetail("value", req.MaxDuration).
					WithDetail("format", "duration string (e.g., 30m, 1h)")
			}
			searchReq.MaxDuration = &maxDuration
		}

		// Execute search
		response, err := search.SearchEpisodes(searchReq)
		if err != nil {
			return appErrors.NewInternalServerError("Failed to search episodes").
				WithDetail("error", err.Error())
		}

		return c.JSON(http.StatusOK, response)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Search podcasts endpoint
	e.GET("/search/podcasts", func(c echo.Context) error {
		var req models.SearchPodcastRequest
		if err := c.Bind(&req); err != nil {
			return appErrors.NewBadRequestError("Invalid query parameters").
				WithDetail("error", err.Error())
		}

		// Build search request
		searchReq := search.PodcastSearchRequest{
			Query:  req.Query,
			Limit:  req.Limit,
			Offset: req.Offset,
		}

		// Execute search
		response, err := search.SearchPodcasts(searchReq)
		if err != nil {
			return appErrors.NewInternalServerError("Failed to search podcasts").
				WithDetail("error", err.Error())
		}

		return c.JSON(http.StatusOK, response)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
}

// registerBatchRoutes registers all batch operation routes
func registerBatchRoutes(e *echo.Echo) {
	// Batch refresh podcasts - refresh metadata for multiple podcasts
	e.POST("/api/batch/refresh", api.BatchRefreshPodcasts, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(5))

	// Batch delete episodes - delete multiple episodes at once
	e.POST("/api/batch/episodes/delete", api.BatchDeleteEpisodes, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(5))

	// Batch add podcasts - add multiple podcasts at once
	e.POST("/api/batch/podcasts/add", api.BatchAddPodcasts, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(5))

	// Get batch job status - check the status of a batch operation
	e.GET("/api/batch/status/:jobId", api.GetBatchJobStatus, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
}

// registerWebhookRoutes registers all webhook configuration routes
func registerWebhookRoutes(e *echo.Echo) {
	// Create webhook configuration
	e.POST("/api/webhooks", func(c echo.Context) error {
		var config models.WebhookConfig
		if err := c.Bind(&config); err != nil {
			return appErrors.NewBadRequestError("Invalid request body").
				WithDetail("error", err.Error())
		}

		// Validate webhook config
		if config.URL == "" {
			return appErrors.NewBadRequestError("URL is required")
		}
		if config.Type == "" {
			return appErrors.NewBadRequestError("Type is required")
		}
		if config.Type != "discord" && config.Type != "slack" && config.Type != "generic" {
			return appErrors.NewBadRequestError("Invalid webhook type").
				WithDetail("allowed", "discord, slack, generic")
		}
		if config.Events == "" {
			return appErrors.NewBadRequestError("Events is required")
		}

		if err := database.SaveWebhookConfig(&config); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to save webhook config")
			return appErrors.NewInternalServerError("Failed to save webhook config")
		}

		return c.JSON(http.StatusCreated, config)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Get all webhook configurations
	e.GET("/api/webhooks", func(c echo.Context) error {
		configs := database.GetAllWebhookConfigs()
		return c.JSON(http.StatusOK, configs)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get specific webhook configuration
	e.GET("/api/webhooks/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return appErrors.NewInvalidParamError("id").WithDetail("value", idStr)
		}

		config := database.GetWebhookConfig(int32(id))
		if config == nil {
			return appErrors.NewNotFoundError("Webhook config not found")
		}

		return c.JSON(http.StatusOK, config)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Update webhook configuration
	e.PUT("/api/webhooks/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return appErrors.NewInvalidParamError("id").WithDetail("value", idStr)
		}

		config := database.GetWebhookConfig(int32(id))
		if config == nil {
			return appErrors.NewNotFoundError("Webhook config not found")
		}

		var updates models.WebhookConfig
		if err := c.Bind(&updates); err != nil {
			return appErrors.NewBadRequestError("Invalid request body").
				WithDetail("error", err.Error())
		}

		// Update fields
		if updates.Name != "" {
			config.Name = updates.Name
		}
		if updates.URL != "" {
			config.URL = updates.URL
		}
		if updates.Type != "" {
			if updates.Type != "discord" && updates.Type != "slack" && updates.Type != "generic" {
				return appErrors.NewBadRequestError("Invalid webhook type").
					WithDetail("allowed", "discord, slack, generic")
			}
			config.Type = updates.Type
		}
		if updates.Events != "" {
			config.Events = updates.Events
		}
		config.Enabled = updates.Enabled

		if err := database.UpdateWebhookConfig(config); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to update webhook config")
			return appErrors.NewInternalServerError("Failed to update webhook config")
		}

		return c.JSON(http.StatusOK, config)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Delete webhook configuration
	e.DELETE("/api/webhooks/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return appErrors.NewInvalidParamError("id").WithDetail("value", idStr)
		}

		if err := database.DeleteWebhookConfig(int32(id)); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to delete webhook config")
			return appErrors.NewInternalServerError("Failed to delete webhook config")
		}

		return c.NoContent(http.StatusNoContent)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))

	// Get webhook delivery history for a specific config
	e.GET("/api/webhooks/:id/deliveries", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return appErrors.NewInvalidParamError("id").WithDetail("value", idStr)
		}

		config := database.GetWebhookConfig(int32(id))
		if config == nil {
			return appErrors.NewNotFoundError("Webhook config not found")
		}

		limitStr := c.QueryParam("limit")
		limit := 50
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}

		deliveries := database.GetWebhookDeliveriesByConfig(int32(id), limit)
		return c.JSON(http.StatusOK, deliveries)
	}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
}

// registerAnalyticsRoutes registers all analytics-related routes
func registerAnalyticsRoutes(e *echo.Echo) {
	analyticsHandler := api.NewAnalyticsHandler()

	// Get popular episodes
	e.GET("/api/analytics/popular", analyticsHandler.GetPopularEpisodes, 
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get episode analytics
	e.GET("/api/analytics/episode/:videoId", analyticsHandler.GetEpisodeAnalytics,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get analytics summary
	e.GET("/api/analytics/summary", analyticsHandler.GetAnalyticsSummary,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get geographic distribution
	e.GET("/api/analytics/geographic", analyticsHandler.GetGeographicDistribution,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get dashboard data (comprehensive analytics)
	e.GET("/api/analytics/dashboard", analyticsHandler.GetDashboardData,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))
}

// registerTranscriptRoutes registers all transcript-related routes
func registerTranscriptRoutes(e *echo.Echo) {
	transcriptHandler := api.NewTranscriptHandler()

	// Get transcript for a video
	e.GET("/api/transcript/:videoId", transcriptHandler.GetTranscript,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get all transcripts for a video
	e.GET("/api/transcript/:videoId/all", transcriptHandler.GetAllTranscripts,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Get available languages for a video's transcripts
	e.GET("/api/transcript/:videoId/languages", transcriptHandler.GetAvailableLanguages,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(20))

	// Explicitly fetch transcript from YouTube
	e.POST("/api/transcript/:videoId/fetch", transcriptHandler.FetchTranscript,
		middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
}
