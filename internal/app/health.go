package app

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sys/unix"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// healthCheckHandler handles the /health endpoint
func healthCheckHandler(c echo.Context) error {
	health := models.HealthResponse{
		Status:     "healthy",
		Components: make(map[string]models.ComponentStatus),
	}

	// Check database connectivity
	dbStatus := checkDatabase()
	health.Components["database"] = dbStatus
	if dbStatus.Status == "down" {
		health.Status = "unhealthy"
	} else if dbStatus.Status == "degraded" && health.Status == "healthy" {
		health.Status = "degraded"
	}

	// Check disk space in audio directory
	diskStatus := checkDiskSpace()
	health.Components["disk_space"] = diskStatus
	if diskStatus.Status == "down" {
		health.Status = "unhealthy"
	} else if diskStatus.Status == "degraded" && health.Status == "healthy" {
		health.Status = "degraded"
	}

	// Check yt-dlp binary exists
	ytdlpStatus := checkYtdlp()
	health.Components["ytdlp"] = ytdlpStatus
	if ytdlpStatus.Status == "down" {
		health.Status = "unhealthy"
	} else if ytdlpStatus.Status == "degraded" && health.Status == "healthy" {
		health.Status = "degraded"
	}

	// Check YouTube API (lightweight check)
	youtubeStatus := checkYouTubeAPI()
	health.Components["youtube_api"] = youtubeStatus
	if youtubeStatus.Status == "down" {
		health.Status = "unhealthy"
	} else if youtubeStatus.Status == "degraded" && health.Status == "healthy" {
		health.Status = "degraded"
	}

	// Return appropriate HTTP status code
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, health)
}

// readinessCheckHandler handles the /ready endpoint for Kubernetes readiness probes
func readinessCheckHandler(c echo.Context) error {
	readiness := models.ReadinessResponse{
		Ready: true,
	}

	// Check critical components for readiness
	// Database must be accessible
	if err := database.CheckHealth(); err != nil {
		readiness.Ready = false
		readiness.Message = "Database not ready: " + err.Error()
		return c.JSON(http.StatusServiceUnavailable, readiness)
	}

	// Audio directory must exist and be writable
	if _, err := os.Stat(config.Config.AudioDir); os.IsNotExist(err) {
		readiness.Ready = false
		readiness.Message = "Audio directory not ready"
		return c.JSON(http.StatusServiceUnavailable, readiness)
	}

	return c.JSON(http.StatusOK, readiness)
}

// checkDatabase checks database connectivity
func checkDatabase() models.ComponentStatus {
	err := database.CheckHealth()
	if err != nil {
		return models.ComponentStatus{
			Status:  "down",
			Message: "Database connection failed: " + err.Error(),
		}
	}
	return models.ComponentStatus{
		Status:  "ok",
		Message: "Database connection healthy",
	}
}

// checkDiskSpace checks available disk space in the audio directory
func checkDiskSpace() models.ComponentStatus {
	var stat unix.Statfs_t
	err := unix.Statfs(config.Config.AudioDir, &stat)
	if err != nil {
		return models.ComponentStatus{
			Status:  "down",
			Message: "Cannot check disk space: " + err.Error(),
		}
	}

	// Calculate available space in GB
	availableGB := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	if availableGB < 1.0 {
		return models.ComponentStatus{
			Status:  "down",
			Message: "Critical: Less than 1GB available",
		}
	} else if availableGB < 5.0 {
		return models.ComponentStatus{
			Status:  "degraded",
			Message: "Warning: Less than 5GB available",
		}
	}

	return models.ComponentStatus{
		Status:  "ok",
		Message: "Disk space healthy",
	}
}

// checkYtdlp checks if yt-dlp binary exists and is executable
func checkYtdlp() models.ComponentStatus {
	// Check if yt-dlp is in PATH
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return models.ComponentStatus{
			Status:  "down",
			Message: "yt-dlp binary not found",
		}
	}

	return models.ComponentStatus{
		Status:  "ok",
		Message: "yt-dlp binary available",
	}
}

// checkYouTubeAPI performs a lightweight check of the YouTube API
func checkYouTubeAPI() models.ComponentStatus {
	if config.Config.GoogleApiKey == "" {
		return models.ComponentStatus{
			Status:  "down",
			Message: "YouTube API key not configured",
		}
	}

	// Create a context with timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithAPIKey(config.Config.GoogleApiKey))
	if err != nil {
		return models.ComponentStatus{
			Status:  "down",
			Message: "Failed to create YouTube service: " + err.Error(),
		}
	}

	// Perform a simple API call to check if the API is accessible
	// We'll just list one video category as a lightweight check
	_, err = service.VideoCategories.List([]string{"snippet"}).RegionCode("US").MaxResults(1).Context(ctx).Do()
	if err != nil {
		return models.ComponentStatus{
			Status:  "degraded",
			Message: "YouTube API call failed: " + err.Error(),
		}
	}

	return models.ComponentStatus{
		Status:  "ok",
		Message: "YouTube API accessible",
	}
}
