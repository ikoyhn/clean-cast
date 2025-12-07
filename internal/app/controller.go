package app

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"ikoyhn/podcast-sponsorblock/internal/services/playlist"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/sponsorblock"
	"net/http"
	"os"
	"path"
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
		rssRequestParams := validateQueryParams(c)
		data := rss.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		if err := checkAuthentication(c); err != nil {
			return err
		}
		validateQueryParams(c)
		data := playlist.BuildPlaylistRssFeed(c.Param("youtubePlaylistId"), handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
		if err := checkAuthentication(c); err != nil {
			return err
		}

		fileName := c.Param("youtubeVideoId")
		if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "..") {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid file name")
		}
		if !common.IsValidParam(fileName) {
			c.Error(echo.NewHTTPError(http.StatusBadRequest, "Invalid channel id"))
		}
		if !common.IsValidFilename(fileName) {
			c.Error(echo.ErrNotFound)
		}

		audioDirAbs, err := filepath.Abs(config.AppConfig.Setup.AudioDir)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server config error")
		}
		requestedFileAbs, err := filepath.Abs(filepath.Join(config.AppConfig.Setup.AudioDir, fileName))
		if err != nil || !strings.HasPrefix(requestedFileAbs, audioDirAbs+string(os.PathSeparator)) && requestedFileAbs != audioDirAbs {
			return echo.ErrNotFound
		}

		file, err := os.Open(requestedFileAbs)
		needRedownload, totalTimeSkipped := sponsorblock.DeterminePodcastDownload(fileName[:len(fileName)-4])
		if file == nil || err != nil || needRedownload {
			database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
			fileName, done := downloader.GetYoutubeVideo(fileName)
			<-done
			file, err = os.Open(path.Join(config.AppConfig.Setup.AudioDir, fileName+".m4a"))
			if err != nil || file == nil {
				return err
			}
			defer file.Close()

			rangeHeader := c.Request().Header.Get("Range")
			if rangeHeader != "" {
				http.ServeFile(c.Response().Writer, c.Request(), path.Join(config.AppConfig.Setup.AudioDir, fileName+".m4a"))
				return nil
			}
			return c.Stream(http.StatusOK, "audio/mp4", file)
		}

		database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			http.ServeFile(c.Response().Writer, c.Request(), path.Join(config.AppConfig.Setup.AudioDir, fileName))
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

func validateQueryParams(c echo.Context) *models.RssRequestParams {
	limitVar := c.Request().URL.Query().Get("limit")
	dateVar := c.Request().URL.Query().Get("date")
	if !common.IsValidParam(c.Param("channelId")) {
		c.Error(echo.NewHTTPError(http.StatusBadRequest, "Invalid channel id"))
	}
	if c.Request().URL.Query().Get("limit") != "" && c.Request().URL.Query().Get("date") != "" {
		c.Error(echo.NewHTTPError(http.StatusBadRequest, "Invalid parameters"))
	}

	if limitVar != "" {
		limitInt, err := strconv.Atoi(c.Request().URL.Query().Get("limit"))
		if err != nil {
			log.Error(err)
			return nil
		}
		return &models.RssRequestParams{Limit: &limitInt, Date: nil}
	}

	if dateVar != "" {
		parsedDate, err := time.Parse("01-02-2006", dateVar)
		if err != nil {
			log.Error("Error parsing date string:", err)
			return nil
		}
		return &models.RssRequestParams{Limit: nil, Date: &parsedDate}
	}
	return &models.RssRequestParams{Limit: nil, Date: nil}
}

func setupCron() {
	cronSchedule := "0 0 * * 0"
	if config.AppConfig.Setup.Cron != "" {
		cronSchedule = config.AppConfig.Setup.Cron
	}
	c := cron.New()
	c.AddFunc(cronSchedule, func() {
		database.DeletePodcastCronJob()
	})
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
