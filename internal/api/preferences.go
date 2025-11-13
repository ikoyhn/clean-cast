package api

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetUserPreferences handles GET /api/preferences
func GetUserPreferences(c echo.Context) error {
	prefs, err := database.GetOrCreateUserPreferences()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get user preferences")
		return appErrors.NewDatabaseError("get_user_preferences", err)
	}

	return c.JSON(http.StatusOK, prefs)
}

// UpdateUserPreferences handles PUT /api/preferences
func UpdateUserPreferences(c echo.Context) error {
	var req struct {
		DefaultQuality    *string `json:"default_quality"`
		DefaultCategories *string `json:"default_categories"`
		AutoDownload      *bool   `json:"auto_download"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return appErrors.NewValidationError("Invalid request body")
	}

	// Get existing preferences
	prefs, err := database.GetOrCreateUserPreferences()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get user preferences")
		return appErrors.NewDatabaseError("get_user_preferences", err)
	}

	// Update only provided fields
	if req.DefaultQuality != nil {
		prefs.DefaultQuality = *req.DefaultQuality
	}
	if req.DefaultCategories != nil {
		prefs.DefaultCategories = *req.DefaultCategories
	}
	if req.AutoDownload != nil {
		prefs.AutoDownload = *req.AutoDownload
	}

	// Save preferences
	if err := database.UpdateUserPreferences(prefs); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update user preferences")
		return appErrors.NewDatabaseError("update_user_preferences", err)
	}

	return c.JSON(http.StatusOK, prefs)
}

// GetFeedPreferences handles GET /api/preferences/feed/:feedId
func GetFeedPreferences(c echo.Context) error {
	feedId := c.Param("feedId")
	if feedId == "" {
		return appErrors.NewInvalidParamError("feedId")
	}

	prefs, err := database.GetFeedPreferences(feedId)
	if err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to get feed preferences")
		return appErrors.NewDatabaseError("get_feed_preferences", err)
	}

	if prefs == nil {
		// Return empty preferences for feed
		return c.JSON(http.StatusOK, map[string]interface{}{
			"feed_id": feedId,
			"message": "No preferences set for this feed",
		})
	}

	return c.JSON(http.StatusOK, prefs)
}

// UpdateFeedPreferences handles PUT /api/preferences/feed/:feedId
func UpdateFeedPreferences(c echo.Context) error {
	feedId := c.Param("feedId")
	if feedId == "" {
		return appErrors.NewInvalidParamError("feedId")
	}

	var req struct {
		FeedType               *string           `json:"feed_type"`
		SponsorBlockCategories *models.StringArray `json:"sponsorblock_categories"`
		MinDuration            *int              `json:"min_duration"`
		MaxDuration            *int              `json:"max_duration"`
		Quality                *string           `json:"quality"`
		AutoDownload           *bool             `json:"auto_download"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return appErrors.NewValidationError("Invalid request body")
	}

	// Validate duration values
	if req.MinDuration != nil && *req.MinDuration < 0 {
		return appErrors.NewValidationError("min_duration must be positive")
	}
	if req.MaxDuration != nil && *req.MaxDuration < 0 {
		return appErrors.NewValidationError("max_duration must be positive")
	}
	if req.MinDuration != nil && req.MaxDuration != nil && *req.MinDuration > *req.MaxDuration {
		return appErrors.NewValidationError("min_duration cannot be greater than max_duration")
	}

	// Get existing preferences or create new
	prefs, err := database.GetFeedPreferences(feedId)
	if err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to get feed preferences")
		return appErrors.NewDatabaseError("get_feed_preferences", err)
	}

	// Create new preferences if none exist
	if prefs == nil {
		prefs = &models.FeedPreferences{
			FeedId:   feedId,
			FeedType: "CHANNEL",
		}
	}

	// Update only provided fields
	if req.FeedType != nil {
		prefs.FeedType = *req.FeedType
	}
	if req.SponsorBlockCategories != nil {
		prefs.SponsorBlockCategories = *req.SponsorBlockCategories
	}
	if req.MinDuration != nil {
		prefs.MinDuration = req.MinDuration
	}
	if req.MaxDuration != nil {
		prefs.MaxDuration = req.MaxDuration
	}
	if req.Quality != nil {
		prefs.Quality = *req.Quality
	}
	if req.AutoDownload != nil {
		prefs.AutoDownload = req.AutoDownload
	}

	// Save preferences
	if err := database.CreateOrUpdateFeedPreferences(prefs); err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to update feed preferences")
		return appErrors.NewDatabaseError("update_feed_preferences", err)
	}

	return c.JSON(http.StatusOK, prefs)
}

// DeleteFeedPreferences handles DELETE /api/preferences/feed/:feedId
func DeleteFeedPreferences(c echo.Context) error {
	feedId := c.Param("feedId")
	if feedId == "" {
		return appErrors.NewInvalidParamError("feedId")
	}

	if err := database.DeleteFeedPreferences(feedId); err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to delete feed preferences")
		return appErrors.NewDatabaseError("delete_feed_preferences", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Feed preferences deleted successfully",
		"feed_id": feedId,
	})
}

// GetAllFeedPreferences handles GET /api/preferences/feeds
func GetAllFeedPreferences(c echo.Context) error {
	prefs, err := database.GetAllFeedPreferences()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get all feed preferences")
		return appErrors.NewDatabaseError("get_all_feed_preferences", err)
	}

	return c.JSON(http.StatusOK, prefs)
}
