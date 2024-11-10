package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/common"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
	"github.com/robfig/cron"
)

func Start() {

	ytdlp.MustInstall(context.TODO(), nil)
	e := echo.New()

	database.ConnectDatabase()
	database.TrackEpisodeFiles()

	setupCron()
	setupLogging(e)
	setupHandlers(e)
	registerRoutes(e)
}

func registerRoutes(e *echo.Echo) {
	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		data := services.BuildRssFeed(c.Param("youtubePlaylistId"), handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})

	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
		fileName := c.Param("youtubeVideoId")
		if !common.IsValidFilename(fileName) {
			c.Error(echo.ErrNotFound)
		}
		file, err := os.Open("/config/audio/" + fileName)
		needRedownload, totalTimeSkipped := services.DeterminePodcastDownload(fileName[:len(fileName)-4])
		if file == nil || err != nil || needRedownload {
			database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
			fileName, done := services.GetYoutubeVideo(fileName)
			<-done
			file, err = os.Open("/config/audio/" + fileName + ".m4a")
			if err != nil || file == nil {
				return err
			}
			defer file.Close()

			rangeHeader := c.Request().Header.Get("Range")
			if rangeHeader != "" {
				http.ServeFile(c.Response().Writer, c.Request(), "/config/audio/"+fileName+".m4a")
				return nil
			}
			return c.Stream(http.StatusOK, "audio/mp4", file)
		}

		database.UpdateEpisodePlaybackHistory(fileName[:len(fileName)-4], totalTimeSkipped)
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			http.ServeFile(c.Response().Writer, c.Request(), "/config/audio/"+fileName)
			return nil
		}
		return c.Stream(http.StatusOK, "audio/mp4", file)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")

	log.Info("Starting server on " + host + ": " + port)
	e.Logger.Fatal(e.Start(host + ":" + port))

}

func setupCron() {
	cronSchedule := "0 0 * * 0"
	if os.Getenv("CRON") != "" {
		cronSchedule = os.Getenv("CRON")
	}
	c := cron.New()
	c.AddFunc(cronSchedule, func() {
		database.DeletePodcastCronJob()
	})
	c.Start()
}

func setupLogging(e *echo.Echo) {
	//custom logging to exclude showing the token from url
	if os.Getenv("TOKEN") != "" {
		logger := middleware.LoggerConfig{
			Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","method":"${method}","path":"${uri.Path}","user_agent":"${user_agent}","status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out}}`,
		}
		e.Use(middleware.LoggerWithConfig(logger))
	} else {
		e.Use(middleware.Logger())
	}
}

func setupHandlers(e *echo.Echo) {
	hostMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if value, ok := os.LookupEnv("TRUSTED_HOSTS"); ok && value != "" {
				log.Info("[AUTH] Checking hosts...")
				host := c.Request().Host
				if !contains(strings.Split(value, ","), host) {
					log.Error("[AUTH] Invalid host")
					return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
				}
			}
			return next(c)
		}
	}

	authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if value, ok := os.LookupEnv("TOKEN"); ok && value != "" {
				log.Info("[AUTH] Checking authentication...")
				authHeader := c.Request().URL.Query().Get("token")
				if authHeader == "" {
					log.Error("[AUTH] Auth not found")
					return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
				}
				if authHeader != value {
					log.Error("[AUTH] Auth not valid")
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
				}
			}
			return next(c)
		}
	}

	if value, ok := os.LookupEnv("TRUSTED_HOSTS"); ok && value != "" {
		e.Use(hostMiddleware)
	}

	if value, ok := os.LookupEnv("TOKEN"); ok && value != "" {
		e.Use(authMiddleware)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
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
