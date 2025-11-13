package api

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/filter"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateFilter handles POST /api/filters
func CreateFilter(c echo.Context) error {
	var req models.Filter

	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return appErrors.NewValidationError("Invalid request body")
	}

	// Validate required fields
	if req.Name == "" {
		return appErrors.NewMissingFieldError("name")
	}
	if req.FilterType == "" {
		return appErrors.NewMissingFieldError("filter_type")
	}

	// Validate the filter
	if err := filter.ValidateFilter(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Filter validation failed")
		return appErrors.NewValidationError(err.Error())
	}

	// Set default action if not provided
	if req.Action == "" {
		req.Action = "block"
	}

	// Create the filter
	if err := database.CreateFilter(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create filter")
		return appErrors.NewDatabaseError("create_filter", err)
	}

	return c.JSON(http.StatusCreated, req)
}

// GetFilters handles GET /api/filters
func GetFilters(c echo.Context) error {
	// Check for feed_id query parameter
	feedId := c.QueryParam("feed_id")

	var filters []models.Filter
	var err error

	if feedId != "" {
		filters, err = database.GetFiltersForFeed(feedId)
	} else {
		filters, err = database.GetAllFilters()
	}

	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get filters")
		return appErrors.NewDatabaseError("get_filters", err)
	}

	return c.JSON(http.StatusOK, filters)
}

// GetFilter handles GET /api/filters/:id
func GetFilter(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return appErrors.NewInvalidParamError("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return appErrors.NewInvalidParamError("id").WithDetail("error", "must be a valid integer")
	}

	filterObj, err := database.GetFilter(int32(id))
	if err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to get filter")
		return appErrors.NewDatabaseError("get_filter", err)
	}

	if filterObj == nil {
		return appErrors.NewResourceNotFoundError("filter", idStr)
	}

	return c.JSON(http.StatusOK, filterObj)
}

// UpdateFilter handles PUT /api/filters/:id
func UpdateFilter(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return appErrors.NewInvalidParamError("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return appErrors.NewInvalidParamError("id").WithDetail("error", "must be a valid integer")
	}

	// Get existing filter
	existing, err := database.GetFilter(int32(id))
	if err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to get filter")
		return appErrors.NewDatabaseError("get_filter", err)
	}

	if existing == nil {
		return appErrors.NewResourceNotFoundError("filter", idStr)
	}

	var req struct {
		Name                   *string `json:"name"`
		FilterType             *string `json:"filter_type"`
		Action                 *string `json:"action"`
		Pattern                *string `json:"pattern"`
		CaseSensitive          *bool   `json:"case_sensitive"`
		ApplyToTitle           *bool   `json:"apply_to_title"`
		ApplyToDescription     *bool   `json:"apply_to_description"`
		FeedId                 *string `json:"feed_id"`
		MinDuration            *int    `json:"min_duration"`
		MaxDuration            *int    `json:"max_duration"`
		Enabled                *bool   `json:"enabled"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return appErrors.NewValidationError("Invalid request body")
	}

	// Update only provided fields
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.FilterType != nil {
		existing.FilterType = *req.FilterType
	}
	if req.Action != nil {
		existing.Action = *req.Action
	}
	if req.Pattern != nil {
		existing.Pattern = *req.Pattern
	}
	if req.CaseSensitive != nil {
		existing.CaseSensitive = *req.CaseSensitive
	}
	if req.ApplyToTitle != nil {
		existing.ApplyToTitle = *req.ApplyToTitle
	}
	if req.ApplyToDescription != nil {
		existing.ApplyToDescription = *req.ApplyToDescription
	}
	if req.FeedId != nil {
		existing.FeedId = req.FeedId
	}
	if req.MinDuration != nil {
		existing.MinDuration = req.MinDuration
	}
	if req.MaxDuration != nil {
		existing.MaxDuration = req.MaxDuration
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	// Validate the updated filter
	if err := filter.ValidateFilter(existing); err != nil {
		logger.Logger.Error().Err(err).Msg("Filter validation failed")
		return appErrors.NewValidationError(err.Error())
	}

	// Update the filter
	if err := database.UpdateFilter(existing); err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to update filter")
		return appErrors.NewDatabaseError("update_filter", err)
	}

	return c.JSON(http.StatusOK, existing)
}

// DeleteFilter handles DELETE /api/filters/:id
func DeleteFilter(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return appErrors.NewInvalidParamError("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return appErrors.NewInvalidParamError("id").WithDetail("error", "must be a valid integer")
	}

	// Check if filter exists
	existing, err := database.GetFilter(int32(id))
	if err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to get filter")
		return appErrors.NewDatabaseError("get_filter", err)
	}

	if existing == nil {
		return appErrors.NewResourceNotFoundError("filter", idStr)
	}

	// Delete the filter
	if err := database.DeleteFilter(int32(id)); err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to delete filter")
		return appErrors.NewDatabaseError("delete_filter", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Filter deleted successfully",
		"id":      id,
	})
}

// ToggleFilter handles PATCH /api/filters/:id/toggle
func ToggleFilter(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return appErrors.NewInvalidParamError("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return appErrors.NewInvalidParamError("id").WithDetail("error", "must be a valid integer")
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to bind request")
		return appErrors.NewValidationError("Invalid request body")
	}

	// Check if filter exists
	existing, err := database.GetFilter(int32(id))
	if err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to get filter")
		return appErrors.NewDatabaseError("get_filter", err)
	}

	if existing == nil {
		return appErrors.NewResourceNotFoundError("filter", idStr)
	}

	// Toggle the filter
	if err := database.ToggleFilter(int32(id), req.Enabled); err != nil {
		logger.Logger.Error().Err(err).Int64("filter_id", id).Msg("Failed to toggle filter")
		return appErrors.NewDatabaseError("toggle_filter", err)
	}

	existing.Enabled = req.Enabled
	return c.JSON(http.StatusOK, existing)
}
