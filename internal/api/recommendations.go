package api

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/middleware"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/recommendations"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRecommendationRoutes registers all recommendation-related routes
func RegisterRecommendationRoutes(e *echo.Echo) {
	db := database.GetDB()
	handler := NewRecommendationsHandler(db)

	// Recommendation API endpoints - no auth required, moderate rate limiting
	api := e.Group("/api/recommendations")
	api.Use(middleware.RateLimitMiddleware(30))

	api.GET("/similar/:podcastId", handler.GetSimilarPodcasts)
	api.GET("/trending", handler.GetTrendingEpisodes)
	api.GET("/related/:videoId", handler.GetRelatedEpisodes)
	api.GET("/for-you", handler.GetPersonalizedRecommendations)
}

// RecommendationsHandler handles recommendation-related HTTP requests
type RecommendationsHandler struct {
	service *recommendations.RecommendationService
}

// NewRecommendationsHandler creates a new recommendations handler
func NewRecommendationsHandler(db *gorm.DB) *RecommendationsHandler {
	return &RecommendationsHandler{
		service: recommendations.NewRecommendationService(db),
	}
}

// GetSimilarPodcasts godoc
// @Summary Get similar podcasts
// @Description Get podcasts similar to the specified podcast based on category and keywords
// @Tags recommendations
// @Accept json
// @Produce json
// @Param podcastId path string true "Podcast ID (YouTube channel or playlist ID)"
// @Param limit query int false "Maximum number of results" default(10)
// @Success 200 {object} map[string]interface{} "Similar podcasts"
// @Failure 400 {object} errors.AppError "Invalid request parameters"
// @Failure 404 {object} errors.AppError "Podcast not found"
// @Failure 500 {object} errors.AppError "Internal server error"
// @Router /api/recommendations/similar/{podcastId} [get]
func (h *RecommendationsHandler) GetSimilarPodcasts(c echo.Context) error {
	podcastId := c.Param("podcastId")
	if !common.IsValidParam(podcastId) {
		return appErrors.NewInvalidParamError("podcastId").
			WithDetail("value", podcastId)
	}

	// Check if podcast exists
	exists, err := database.PodcastExists(podcastId)
	if err != nil {
		logger.Logger.Error().Err(err).Str("podcast_id", podcastId).Msg("Failed to check podcast existence")
		return appErrors.NewDatabaseError("check_podcast_exists", err)
	}
	if !exists {
		return appErrors.NewResourceNotFoundError("podcast", podcastId)
	}

	// Get limit parameter
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 50 {
			return appErrors.NewInvalidParamError("limit").
				WithDetail("value", limitStr).
				WithDetail("allowed_range", "1-50")
		}
		limit = parsedLimit
	}

	// Get similar podcasts
	similar, err := h.service.GetSimilarPodcasts(podcastId, limit)
	if err != nil {
		logger.Logger.Error().Err(err).Str("podcast_id", podcastId).Msg("Failed to get similar podcasts")
		return appErrors.NewInternalError("Failed to retrieve similar podcasts")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"podcast_id": podcastId,
		"limit":      limit,
		"results":    len(similar),
		"similar":    similar,
	})
}

// GetTrendingEpisodes godoc
// @Summary Get trending episodes
// @Description Get the most played episodes in the last 7 days
// @Tags recommendations
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results" default(10)
// @Success 200 {object} map[string]interface{} "Trending episodes"
// @Failure 400 {object} errors.AppError "Invalid request parameters"
// @Failure 500 {object} errors.AppError "Internal server error"
// @Router /api/recommendations/trending [get]
func (h *RecommendationsHandler) GetTrendingEpisodes(c echo.Context) error {
	// Get limit parameter
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 50 {
			return appErrors.NewInvalidParamError("limit").
				WithDetail("value", limitStr).
				WithDetail("allowed_range", "1-50")
		}
		limit = parsedLimit
	}

	// Get trending episodes
	trending, err := h.service.GetTrendingEpisodes(limit)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get trending episodes")
		return appErrors.NewInternalError("Failed to retrieve trending episodes")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"period":   "last_7_days",
		"limit":    limit,
		"results":  len(trending),
		"episodes": trending,
	})
}

// GetRelatedEpisodes godoc
// @Summary Get related episodes
// @Description Get episodes from the same channel/playlist as the specified video
// @Tags recommendations
// @Accept json
// @Produce json
// @Param videoId path string true "YouTube Video ID"
// @Param limit query int false "Maximum number of results" default(10)
// @Success 200 {object} map[string]interface{} "Related episodes"
// @Failure 400 {object} errors.AppError "Invalid request parameters"
// @Failure 404 {object} errors.AppError "Episode not found"
// @Failure 500 {object} errors.AppError "Internal server error"
// @Router /api/recommendations/related/{videoId} [get]
func (h *RecommendationsHandler) GetRelatedEpisodes(c echo.Context) error {
	videoId := c.Param("videoId")
	if !common.IsValidParam(videoId) {
		return appErrors.NewInvalidParamError("videoId").
			WithDetail("value", videoId)
	}

	// Get limit parameter
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 50 {
			return appErrors.NewInvalidParamError("limit").
				WithDetail("value", limitStr).
				WithDetail("allowed_range", "1-50")
		}
		limit = parsedLimit
	}

	// Get related episodes
	related, err := h.service.GetRelatedEpisodes(videoId, limit)
	if err != nil {
		logger.Logger.Error().Err(err).Str("video_id", videoId).Msg("Failed to get related episodes")
		// Check if it's a not found error
		if err == gorm.ErrRecordNotFound {
			return appErrors.NewResourceNotFoundError("episode", videoId)
		}
		return appErrors.NewInternalError("Failed to retrieve related episodes")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"video_id": videoId,
		"limit":    limit,
		"results":  len(related),
		"episodes": related,
	})
}

// GetPersonalizedRecommendations godoc
// @Summary Get personalized recommendations
// @Description Get personalized episode recommendations based on listening history
// @Tags recommendations
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results" default(10)
// @Success 200 {object} map[string]interface{} "Personalized recommendations"
// @Failure 400 {object} errors.AppError "Invalid request parameters"
// @Failure 500 {object} errors.AppError "Internal server error"
// @Router /api/recommendations/for-you [get]
func (h *RecommendationsHandler) GetPersonalizedRecommendations(c echo.Context) error {
	// Get limit parameter
	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 50 {
			return appErrors.NewInvalidParamError("limit").
				WithDetail("value", limitStr).
				WithDetail("allowed_range", "1-50")
		}
		limit = parsedLimit
	}

	// Get personalized recommendations
	recommendations, err := h.service.GetPersonalizedRecommendations(limit)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get personalized recommendations")
		return appErrors.NewInternalError("Failed to retrieve personalized recommendations")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"strategy": "personalized_based_on_history",
		"limit":    limit,
		"results":  len(recommendations),
		"episodes": recommendations,
	})
}
