package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/middleware"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lrstanley/go-ytdlp"
	"github.com/robfig/cron"
)

var (
	activeDownloads sync.WaitGroup
	cronJob         *cron.Cron
)

// AddActiveDownload increments the active download counter
func AddActiveDownload() {
	activeDownloads.Add(1)
}

// DoneActiveDownload decrements the active download counter
func DoneActiveDownload() {
	activeDownloads.Done()
}

func Start() {
	ytdlp.MustInstall(context.TODO(), nil)

	// Setup download tracking callbacks
	downloader.SetDownloadCallbacks(AddActiveDownload, DoneActiveDownload)

	e := echo.New()
	e.HideBanner = true
	setupLogging(e)

	// Setup custom error handler
	e.HTTPErrorHandler = middleware.ErrorHandler()

	// Register middleware (order matters!)
	// 1. Request ID - must be first so it's available to all other middleware
	e.Use(middleware.RequestIDMiddleware())

	// 2. Recover from panics
	e.Use(middleware.RecoverMiddleware())

	// 3. Validation middleware
	e.Use(middleware.ValidationErrorMiddleware())

	database.SetupDatabase()
	database.TrackEpisodeFiles()

	cronJob = setupCron()

	setupHandlers(e)

	// Start server in a goroutine
	go func() {
		registerRoutes(e)
	}()

	// Setup graceful shutdown
	setupGracefulShutdown(e)
}

// setupGracefulShutdown handles SIGTERM and SIGINT signals for graceful shutdown
func setupGracefulShutdown(e *echo.Echo) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	logger.Logger.Info().Msg("Shutting down server gracefully...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)
	defer cancel()

	// Stop the cron job
	if cronJob != nil {
		logger.Logger.Info().Msg("Stopping cron job...")
		cronJob.Stop()
	}

	// Stop accepting new requests and wait for active downloads
	logger.Logger.Info().Msg("Waiting for active downloads to complete (max 30 seconds)...")

	// Wait for active downloads with timeout
	downloadsDone := make(chan struct{})
	go func() {
		activeDownloads.Wait()
		close(downloadsDone)
	}()

	select {
	case <-downloadsDone:
		logger.Logger.Info().Msg("All active downloads completed")
	case <-ctx.Done():
		logger.Logger.Warn().Msg("Shutdown timeout reached, some downloads may not have completed")
	}

	// Close database connections
	logger.Logger.Info().Msg("Closing database connections...")
	if err := database.Close(); err != nil {
		logger.Logger.Error().Err(err).Msg("Error closing database")
	}

	// Shutdown HTTP server
	logger.Logger.Info().Msg("Shutting down HTTP server...")
	if err := e.Shutdown(ctx); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Logger.Info().Msg("Server shutdown complete")
}
