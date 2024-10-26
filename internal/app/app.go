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

	e.Use(middleware.Logger())

	env.registerRoutes(e)
}

func (env *Env) registerRoutes(e *echo.Echo) {
	e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
		data := services.BuildRssFeed(env.db, c, c.Param("youtubePlaylistId"))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	})
	e.GET("/media/:youtubeVideoId", func(c echo.Context) error {
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

		fileBytes := make([]byte, fileInfo.Size())
		_, err = file.Read(fileBytes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read file")
		}

		c.Response().Header().Set("Connection", "close")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", fileName+".m4a"))
		c.Response().Header().Set("Content-Type", "audio/mp4")
		c.Response().Header().Set("Content-Length", strconv.Itoa(int(fileInfo.Size())))
		c.Response().Header().Set("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("ETag", strconv.FormatInt(fileInfo.ModTime().UnixNano(), 10))
		return c.Stream(http.StatusOK, "audio/mp4", bytes.NewReader(fileBytes))
	})
	e.HEAD("/media/:youtubeVideoId", func(c echo.Context) error {
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
		c.Response().Header().Set("ETag", strconv.FormatInt(fileInfo.ModTime().UnixNano(), 10))
		c.Response().Header().Set("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
		return c.NoContent(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")

	log.Info("Starting server on " + host + ": " + port)
	e.Logger.Fatal(e.Start(host + ":" + port))

}
