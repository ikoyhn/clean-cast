package app

import (
	"bytes"
	"context"
	"encoding/base64"
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

	e.Use(middleware.Logger())

	env.registerRoutes(e)
}

func (env *Env) registerRoutes(e *echo.Echo) {
	// Define the trusted hosts
	trustedHosts := strings.Split(os.Getenv("TRUSTED_HOSTS"), ",")

	// Define the authentication credentials
	// usernameEnv := os.Getenv("USERNAME")
	passwordEnv := os.Getenv("PASSWORD")

	// Create a custom middleware to check the request's Host header
	middlewareFunc := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if trustedHosts != nil {
				log.Info("[Security] Checking host...")
				log.Info("[Security] Host: " + c.Request().Host)
				log.Info("[Security] Trusted hosts: " + strings.Join(trustedHosts, ", "))
				host := c.Request().Host
				if !contains(trustedHosts, host) {
					log.Error("[Security] Invalid Host!")
					return echo.NewHTTPError(http.StatusForbidden, "Host not trusted")
				}
			}
			log.Info("[Security] Host validated")
			return next(c)
		}
	}
	e.Use(middleware.BasicAuth(middleware.BasicAuthValidator(func(username, password string, c echo.Context) (bool, error) {
		// Check if Basic Auth header is present and valid
		auth := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			return false, nil
		}
		decoded, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			return false, nil
		}
		decodedPassword := strings.Split(string(decoded), ":")[1]
		if decodedPassword != passwordEnv {
			return false, fmt.Errorf("invalid credentials")
		}
		return true, nil
	})))

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			if httpErr.Code == http.StatusUnauthorized {
				c.Response().Header().Set("WWW-Authenticate", "Basic realm=\"podcastsponsorblock\"")
				c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
				c.Response().Header().Set("Content-Length", "0")
				c.Response().Header().Set("Connection", "close")
				c.Response().WriteHeader(http.StatusUnauthorized)
				c.Response().Write([]byte{})
				return
			}
		}
		c.Logger().Error(err)
		c.Response().WriteHeader(http.StatusInternalServerError)
		c.Response().Write([]byte("Internal Server Error"))
	}

	e.GET("/rss/:youtubePlaylistId", middlewareFunc(func(c echo.Context) error {
		data := services.BuildRssFeed(env.db, c.Param("youtubePlaylistId"), handler(c.Request()))
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
		c.Response().Header().Del("Transfer-Encoding")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
	}))

	e.Match([]string{"GET", "HEAD"}, "/media/:youtubeVideoId", middlewareFunc(func(c echo.Context) error {
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
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")

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
