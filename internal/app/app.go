package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"

	"github.com/labstack/echo/v4"
	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
)

func Start() {

	if _, err := config.Load(); err != nil {
		panic(err)
	}
	youtube.SetupYoutubeService()

	ctx := context.TODO()
	if _, err := ytdlp.Install(ctx, &ytdlp.InstallOptions{AllowVersionMismatch: true}); err != nil {
		panic(err)
	}
	if _, err := ytdlp.New().Update(ctx); err != nil {
		log.Warnf("yt-dlp self-update failed, continuing with current version: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	setupLogging(e)

	database.SetupDatabase()
	database.TrackEpisodeFiles()

	setupCron()

	setupHandlers(e)
	registerRoutes(e)
}
