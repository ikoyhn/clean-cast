package app

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/channel"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"ikoyhn/podcast-sponsorblock/internal/services/playlist"
	"ikoyhn/podcast-sponsorblock/internal/services/sponsorblock"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/labstack/gommon/log"
	"github.com/robfig/cron"
)

func registerRoutes(e *echo.Echo) {
	e.GET("/channel/:channelId", func(c echo.Context) error {
		if err := checkAuthentication(c); err != nil {
			return err
		}
		rssRequestParams, err := validateQueryParams(c)
		if err != nil {
			return err
		}
		data := channel.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()))
		if data == nil {
			return echo.NewHTTPError(http.StatusBadGateway, "Could not build channel feed")
		}
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		if err := checkAuthentication(c); err != nil {
			return err
		}
		if _, err := validateQueryParams(c); err != nil {
			return err
		}
		playlistId := strings.Split(c.Param("youtubePlaylistId"), "&")[0]
		data := playlist.BuildPlaylistRssFeed(playlistId, handler(c.Request()))
		if data == nil {
			return echo.NewHTTPError(http.StatusBadGateway, "Could not build playlist feed")
		}
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
		if err := checkAuthentication(c); err != nil {
			return err
		}

		youtubeVideoId := c.Param("youtubeVideoId")
		if strings.Contains(youtubeVideoId, "/") || strings.Contains(youtubeVideoId, "\\") || strings.Contains(youtubeVideoId, "..") {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid file name")
		}
		if !common.IsValidParam(youtubeVideoId) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid video id")
		}

		audioDirAbs, err := filepath.Abs(config.AppConfig.Setup.AudioDir)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server config error")
		}

		needRedownload, totalTimeSkipped := sponsorblock.DeterminePodcastDownload(youtubeVideoId)
		database.UpdateEpisodePlaybackHistory(youtubeVideoId, totalTimeSkipped)

		filePath := database.FindFileWithId(audioDirAbs, youtubeVideoId)
		file, err := os.Open(filePath)

		if file == nil || err != nil || needRedownload {
			if file != nil {
				file.Close()
			}
			<-downloader.GetYoutubeVideo(youtubeVideoId, needRedownload)
			filePath = database.FindFileWithId(audioDirAbs, youtubeVideoId)
			file, err = os.Open(filePath)
			if err != nil || file == nil {
				log.Errorf("[MEDIA] No file available for %s: %v", youtubeVideoId, err)
				return echo.NewHTTPError(http.StatusNotFound, "Episode unavailable")
			}
		}

		defer file.Close()
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			http.ServeFile(c.Response().Writer, c.Request(), filePath)
			return nil
		}
		return c.Stream(http.StatusOK, "audio/mp4", file)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	host := os.Getenv("HOST")

	log.Debug("Starting server on " + host + ": " + port)
	e.Logger.Fatal(e.Start(host + ":" + port))

}

func validateQueryParams(c echo.Context) (*models.RssRequestParams, error) {
	limitVar := c.Request().URL.Query().Get("limit")
	dateVar := c.Request().URL.Query().Get("date")
	if !common.IsValidParam(c.Param("channelId")) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid channel id")
	}
	if limitVar != "" && dateVar != "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid parameters")
	}

	if limitVar != "" {
		limitInt, err := strconv.Atoi(limitVar)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid limit")
		}
		return &models.RssRequestParams{Limit: &limitInt, Date: nil}, nil
	}

	if dateVar != "" {
		parsedDate, err := time.Parse("01-02-2006", dateVar)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid date, expected MM-DD-YYYY")
		}
		return &models.RssRequestParams{Limit: nil, Date: &parsedDate}, nil
	}
	return &models.RssRequestParams{Limit: nil, Date: nil}, nil
}

func setupCron() {
	cronSchedule := "0 0 * * 0"
	if config.AppConfig.Setup.Cron != "" {
		cronSchedule = config.AppConfig.Setup.Cron
	}

	schedule, err := cron.ParseStandard(cronSchedule)
	if err != nil {
		log.Errorf("[CRON] Invalid cron schedule %q (%v), falling back to weekly", cronSchedule, err)
		schedule, _ = cron.ParseStandard("0 0 * * 0")
	}
	c := cron.New()
	c.Schedule(schedule, cron.FuncJob(database.DeletePodcastCronJob))
	c.Start()
}

func setupHandlers(e *echo.Echo) {
	hostMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if value, ok := os.LookupEnv("TRUSTED_HOSTS"); ok && value != "" {
				log.Info("[AUTH] Checking hosts...")
				host := c.Request().Host
				if !common.Contains(strings.Split(value, ","), host) {
					log.Error("[AUTH] Invalid host")
					return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
				}
			}
			return next(c)
		}
	}

	if value, ok := os.LookupEnv("TRUSTED_HOSTS"); ok && value != "" {
		e.Use(hostMiddleware)
	}
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

func checkAuthentication(c echo.Context) error {
	tokenConfigured := config.AppConfig.Authentication.Token != ""
	basicConfigured := config.AppConfig.Authentication.BasicAuth.Password != ""

	// If no authentication configured, allow through
	if !tokenConfigured && !basicConfigured {
		return nil
	}

	// If both basic and token are configured, accept either method (Basic OR token)
	if basicConfigured && tokenConfigured {
		user, pass, ok := c.Request().BasicAuth()
		token := c.Request().URL.Query().Get("token")

		basicOk := ok && pass == config.AppConfig.Authentication.BasicAuth.Password
		if config.AppConfig.Authentication.BasicAuth.Username != "" {
			basicOk = basicOk && user == config.AppConfig.Authentication.BasicAuth.Username
		}

		tokenOk := token == config.AppConfig.Authentication.Token

		if basicOk || tokenOk {
			return nil
		}

		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	if basicConfigured {
		_, pass, ok := c.Request().BasicAuth()
		if ok && pass == config.AppConfig.Authentication.BasicAuth.Password {
			return nil
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	if tokenConfigured {
		token := c.Request().URL.Query().Get("token")
		if token == config.AppConfig.Authentication.Token {
			return nil
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}
