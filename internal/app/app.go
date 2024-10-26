package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
	"gorm.io/gorm"
	setup "ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	db *gorm.DB
}

func Start() {

	ytdlp.MustInstall(context.TODO(), nil)

	db, err := setup.ConnectDatabase()
	if err != nil {
		log.Panic(err)
	}
	env := &Env{db: db}

	e := echo.New()

	// Create a custom logger middleware
	setupLogging(e)
	env.registerRoutes(e)
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

func (env *Env) registerRoutes(e *echo.Echo) {
	trustedHosts := strings.Split(os.Getenv("TRUSTED_HOSTS"), ",")
	authKey := os.Getenv("TOKEN")

	// Create a custom middleware to check the request's Host header
	hostMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if trustedHosts != nil {
				log.Info("[AUTH] Checking authentication...")
				host := c.Request().Host
				if !contains(trustedHosts, host) {
					log.Error("[AUTH] Invalid host")
					return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
				}
			}
			return next(c)
		}
	}

	authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if authKey != "" {
				log.Info("[AUTH] Checking authentication...")
				authHeader := c.Request().URL.Query().Get("token")
				if authHeader == "" {
					log.Error("[AUTH] Auth not found")
					return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
				}
				if authHeader != authKey {
					log.Error("[AUTH] Auth not valid")
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
				}
			}
			return next(c)
		}
	}

	e.GET("/rss/:youtubePlaylistId", authMiddleware(hostMiddleware(func(c echo.Context) error {
		data := services.BuildRssFeed(env.db, c.Param("youtubePlaylistId"), handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})))

	e.Match([]string{"GET", "HEAD"}, "/media/:youtubeVideoId", authMiddleware(hostMiddleware(func(c echo.Context) error {
		fileName, done := services.GetYoutubeVideo(c.Param("youtubeVideoId"))
		<-done

		file, err := os.Open("/config/" + fileName + ".m4a")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open file: /config/"+fileName+".m4a")
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve file info")
		}

		c.Response().Header().Set("Connection", "close")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", fileName+".m4a"))
		c.Response().Header().Set("Content-Type", "audio/mp4")
		c.Response().Header().Set("Content-Length", strconv.Itoa(int(fileInfo.Size())))
		c.Response().Header().Set("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("ETag", strconv.FormatInt(fileInfo.ModTime().UnixNano(), 10))

		if c.Request().Method == "HEAD" {
			return nil
		}

		fileBytes := make([]byte, fileInfo.Size())
		_, err = file.Read(fileBytes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read file")
		}

		return c.Stream(http.StatusOK, "audio/mp4", bytes.NewReader(fileBytes))
	})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")

	log.Info("Starting server on " + host + ": " + port)
	e.Logger.Fatal(e.Start(host + ":" + port))

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
