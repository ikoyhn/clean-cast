package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"

	"github.com/labstack/echo/v4"
	"github.com/lrstanley/go-ytdlp"
)

func Start() {

	if _, err := config.Load(); err != nil {
		panic(err)
	}
	youtube.SetupYoutubeService()
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
