package api

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/services/analytics"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	service *analytics.Service
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		service: analytics.NewService(),
	}
}

// GetPopularEpisodes handles GET /api/analytics/popular
// Query params: limit (default: 10), days (default: 7)
func (h *AnalyticsHandler) GetPopularEpisodes(c echo.Context) error {
	// Parse query parameters
	limitStr := c.QueryParam("limit")
	daysStr := c.QueryParam("days")

	limit := 10
	days := 7

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Logger.Warn().
				Err(err).
				Str("limit", limitStr).
				Msg("Invalid limit parameter")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid limit parameter",
			})
		}
		if parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err != nil {
			logger.Logger.Warn().
				Err(err).
				Str("days", daysStr).
				Msg("Invalid days parameter")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter",
			})
		}
		if parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	logger.Logger.Debug().
		Int("limit", limit).
		Int("days", days).
		Msg("Fetching popular episodes")

	episodes, err := h.service.GetPopularEpisodes(limit, days)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching popular episodes")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch popular episodes",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"limit":    limit,
		"days":     days,
		"episodes": episodes,
	})
}

// GetEpisodeAnalytics handles GET /api/analytics/episode/:videoId
func (h *AnalyticsHandler) GetEpisodeAnalytics(c echo.Context) error {
	videoId := c.Param("videoId")

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video ID is required",
		})
	}

	logger.Logger.Debug().
		Str("video_id", videoId).
		Msg("Fetching episode analytics")

	analyticsData, err := h.service.GetEpisodeAnalytics(videoId)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Error fetching episode analytics")
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Analytics not found for this episode",
		})
	}

	return c.JSON(http.StatusOK, analyticsData)
}

// GetAnalyticsSummary handles GET /api/analytics/summary
func (h *AnalyticsHandler) GetAnalyticsSummary(c echo.Context) error {
	logger.Logger.Debug().Msg("Fetching analytics summary")

	summary, err := h.service.GetAnalyticsSummary()
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching analytics summary")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch analytics summary",
		})
	}

	return c.JSON(http.StatusOK, summary)
}

// GetGeographicDistribution handles GET /api/analytics/geographic
// Query params: limit (default: 10)
func (h *AnalyticsHandler) GetGeographicDistribution(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 10

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Logger.Warn().
				Err(err).
				Str("limit", limitStr).
				Msg("Invalid limit parameter")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid limit parameter",
			})
		}
		if parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	logger.Logger.Debug().
		Int("limit", limit).
		Msg("Fetching geographic distribution")

	distribution, err := h.service.GetGeographicDistribution(limit)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching geographic distribution")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch geographic distribution",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"limit":        limit,
		"distribution": distribution,
	})
}

// GetDashboardData handles GET /api/analytics/dashboard
// Returns comprehensive analytics for a dashboard view
func (h *AnalyticsHandler) GetDashboardData(c echo.Context) error {
	logger.Logger.Debug().Msg("Fetching dashboard data")

	// Fetch summary
	summary, err := h.service.GetAnalyticsSummary()
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching analytics summary for dashboard")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch dashboard data",
		})
	}

	// Fetch popular episodes
	popularEpisodes, err := h.service.GetPopularEpisodes(10, 7)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching popular episodes for dashboard")
		popularEpisodes = []interface{}{} // Use empty array on error
	}

	// Fetch geographic distribution
	geographic, err := h.service.GetGeographicDistribution(10)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Error fetching geographic distribution for dashboard")
		geographic = []map[string]interface{}{} // Use empty array on error
	}

	dashboardData := map[string]interface{}{
		"summary":                 summary,
		"popular_episodes":        popularEpisodes,
		"geographic_distribution": geographic,
	}

	return c.JSON(http.StatusOK, dashboardData)
}
