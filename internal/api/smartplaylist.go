package api

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/smartplaylist"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CreateSmartPlaylistHandler creates a new smart playlist
func CreateSmartPlaylistHandler(c echo.Context) error {
	logger.Logger.Info().Msg("[SMART PLAYLIST] Create request received")

	var req models.SmartPlaylistCreateRequest
	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name is required",
		})
	}

	if len(req.Rules) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "At least one rule is required",
		})
	}

	// Set default logic if not provided
	if req.Logic == "" {
		req.Logic = "AND"
	}

	// Validate logic
	if req.Logic != "AND" && req.Logic != "OR" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Logic must be either 'AND' or 'OR'",
		})
	}

	// Create smart playlist
	playlist, err := smartplaylist.CreateSmartPlaylist(&req)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create smart playlist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create smart playlist",
		})
	}

	// Convert to response
	response, err := smartplaylist.GetSmartPlaylist(playlist.Id)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get created smart playlist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve created smart playlist",
		})
	}

	logger.Logger.Info().Str("id", playlist.Id).Str("name", playlist.Name).Msg("[SMART PLAYLIST] Created successfully")

	return c.JSON(http.StatusCreated, response)
}

// GetSmartPlaylistHandler retrieves a smart playlist by ID
func GetSmartPlaylistHandler(c echo.Context) error {
	id := c.Param("id")
	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] Get request received")

	playlist, err := smartplaylist.GetSmartPlaylist(id)
	if err != nil {
		logger.Logger.Error().Err(err).Str("id", id).Msg("Failed to get smart playlist")
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Smart playlist not found",
		})
	}

	return c.JSON(http.StatusOK, playlist)
}

// GetAllSmartPlaylistsHandler retrieves all smart playlists
func GetAllSmartPlaylistsHandler(c echo.Context) error {
	logger.Logger.Info().Msg("[SMART PLAYLIST] Get all request received")

	playlists, err := smartplaylist.GetAllSmartPlaylists()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get smart playlists")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve smart playlists",
		})
	}

	return c.JSON(http.StatusOK, playlists)
}

// UpdateSmartPlaylistHandler updates an existing smart playlist
func UpdateSmartPlaylistHandler(c echo.Context) error {
	id := c.Param("id")
	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] Update request received")

	var req models.SmartPlaylistUpdateRequest
	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Validate logic if provided
	if req.Logic != "" && req.Logic != "AND" && req.Logic != "OR" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Logic must be either 'AND' or 'OR'",
		})
	}

	// Update smart playlist
	playlist, err := smartplaylist.UpdateSmartPlaylist(id, &req)
	if err != nil {
		logger.Logger.Error().Err(err).Str("id", id).Msg("Failed to update smart playlist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update smart playlist",
		})
	}

	// Convert to response
	response, err := smartplaylist.GetSmartPlaylist(playlist.Id)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get updated smart playlist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve updated smart playlist",
		})
	}

	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] Updated successfully")

	return c.JSON(http.StatusOK, response)
}

// DeleteSmartPlaylistHandler deletes a smart playlist
func DeleteSmartPlaylistHandler(c echo.Context) error {
	id := c.Param("id")
	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] Delete request received")

	err := smartplaylist.DeleteSmartPlaylist(id)
	if err != nil {
		logger.Logger.Error().Err(err).Str("id", id).Msg("Failed to delete smart playlist")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete smart playlist",
		})
	}

	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] Deleted successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Smart playlist deleted successfully",
	})
}

// GetSmartPlaylistRSSHandler generates RSS feed for a smart playlist
func GetSmartPlaylistRSSHandler(c echo.Context) error {
	id := c.Param("id")
	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] RSS feed request received")

	// Get the host from the request
	var scheme string
	if c.Request().TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	host := scheme + "://" + c.Request().Host

	// Generate RSS feed
	rssFeed, err := smartplaylist.BuildSmartPlaylistRSSFeed(id, host)
	if err != nil {
		logger.Logger.Error().Err(err).Str("id", id).Msg("Failed to generate RSS feed")
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Smart playlist not found or failed to generate RSS feed",
		})
	}

	c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(rssFeed)))

	logger.Logger.Info().Str("id", id).Msg("[SMART PLAYLIST] RSS feed generated successfully")

	return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", rssFeed)
}
