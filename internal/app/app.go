package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/database"

	"github.com/labstack/echo/v4"
	"github.com/lrstanley/go-ytdlp"
)

func Start() {
	ytdlp.MustInstall(context.TODO(), nil)

	e := echo.New()
	e.HideBanner = true
	setupLogging(e)

	database.SetupDatabase()
	database.TrackEpisodeFiles()

	setupCron()

	setupHandlers(e)
	registerRoutes(e)
}
