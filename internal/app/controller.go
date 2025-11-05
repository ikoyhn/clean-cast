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
	"strconv"

	"time"

	"github.com/labstack/echo/v4"
	log "github.com/labstack/gommon/log"
	"github.com/robfig/cron"
)

func registerRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

e.GET("/channel/:channelId", func(c echo.Context) error {
		
		rssRequestParams := validateQueryParams(c)
		data := rss.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		validateQueryParams(c)
		data := playlist.BuildPlaylistRssFeed(c.Param("youtubePlaylistId"), handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
				fileName := c.Param("youtubeVideoId")
		if !common.IsValidParam(fileName) {
			c.Error(echo.NewHTTPError(http.StatusBadRequest, "Invalid channel id"))
		}
		if !common.IsValidFilename(fileName) {
			c.Error(echo.ErrNotFound)
		}

		file, err := os.Open(path.Join(config.Config.AudioDir, fileName))
		needRedownload, totalTimeSkipped := sponsorblock.DeterminePodcastDownload(fileName[:len(fileName)-4])
		if file == nil || err != nil || needRedownload {
			database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
			fileName, done := downloader.GetYoutubeVideo(fileName)
			<-done
			file, err = os.Open(path.Join(config.Config.AudioDir, fileName+".m4a"))
			if err != nil || file == nil {
				return err
			}
			defer file.Close()

			rangeHeader := c.Request().Header.Get("Range")
			if rangeHeader != "" {
				http.ServeFile(c.Response().Writer, c.Request(), path.Join(config.Config.AudioDir, fileName+".m4a"))
				return nil
			}
			return c.Stream(http.StatusOK, "audio/mp4", file)
		}

		database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			http.ServeFile(c.Response().Writer, c.Request(), path.Join(config.Config.AudioDir, fileName))
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
	if config.Config.Cron != "" {
		cronSchedule = config.Config.Cron
	}
	c := cron.New()
	c.AddFunc(cronSchedule, func() {
		database.DeletePodcastCronJob()
	})
	c.Start()
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
