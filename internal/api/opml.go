package api

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/opml"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ImportOPMLHandler handles OPML import requests
func ImportOPMLHandler(c echo.Context) error {
	logger.Logger.Info().Msg("[OPML] Import request received")

	// Get file from multipart form
	file, err := c.FormFile("file")
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get file from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No file provided or invalid file format",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to open uploaded file")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process uploaded file",
		})
	}
	defer src.Close()

	// Read file contents
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to read file contents")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to read file contents",
		})
	}

	// Import OPML
	response := opml.ImportOPML(fileBytes)

	logger.Logger.Info().
		Int("success", response.Success).
		Int("failed", response.Failed).
		Int("total", response.TotalFeeds).
		Msg("[OPML] Import completed")

	return c.JSON(http.StatusOK, response)
}

// ExportOPMLHandler handles OPML export requests
func ExportOPMLHandler(c echo.Context) error {
	logger.Logger.Info().Msg("[OPML] Export request received")

	// Get the host from the request
	var scheme string
	if c.Request().TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	host := scheme + "://" + c.Request().Host

	// Generate OPML
	opmlData, err := opml.ExportOPML(host)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to export OPML")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate OPML export",
		})
	}

	// Set headers for file download
	filename := fmt.Sprintf("cleancast-subscriptions-%s.opml", time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(opmlData)))

	logger.Logger.Info().Str("filename", filename).Msg("[OPML] Export completed")

	return c.Blob(http.StatusOK, "application/xml; charset=utf-8", opmlData)
}
