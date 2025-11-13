package search

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"time"
)

// EpisodeSearchRequest holds the request parameters for episode search
type EpisodeSearchRequest struct {
	Query       string
	PodcastId   *string
	StartDate   *time.Time
	EndDate     *time.Time
	MinDuration *time.Duration
	MaxDuration *time.Duration
	Type        *string
	Limit       int
	Offset      int
}

// PodcastSearchRequest holds the request parameters for podcast search
type PodcastSearchRequest struct {
	Query  string
	Limit  int
	Offset int
}

// EpisodeSearchResponse holds the response for episode search
type EpisodeSearchResponse struct {
	Episodes   []models.PodcastEpisode `json:"episodes"`
	TotalCount int64                   `json:"total_count"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
	HasMore    bool                    `json:"has_more"`
}

// PodcastSearchResponse holds the response for podcast search
type PodcastSearchResponse struct {
	Podcasts   []models.Podcast `json:"podcasts"`
	TotalCount int64            `json:"total_count"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	HasMore    bool             `json:"has_more"`
}

// SearchEpisodes searches for episodes based on the provided parameters
func SearchEpisodes(req EpisodeSearchRequest) (*EpisodeSearchResponse, error) {
	logger.Logger.Debug().
		Str("query", req.Query).
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Msg("Searching episodes")

	// Set default limit if not specified
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Cap maximum limit
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Build database search parameters
	params := database.SearchEpisodesParams{
		Query:       req.Query,
		PodcastId:   req.PodcastId,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		MinDuration: req.MinDuration,
		MaxDuration: req.MaxDuration,
		Type:        req.Type,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	// Execute search
	episodes, totalCount, err := database.SearchEpisodes(params)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("query", req.Query).
			Msg("Failed to search episodes")
		return nil, err
	}

	// Calculate if there are more results
	hasMore := int64(req.Offset+req.Limit) < totalCount

	response := &EpisodeSearchResponse{
		Episodes:   episodes,
		TotalCount: totalCount,
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    hasMore,
	}

	logger.Logger.Debug().
		Int64("total_count", totalCount).
		Int("results", len(episodes)).
		Bool("has_more", hasMore).
		Msg("Episode search completed")

	return response, nil
}

// SearchPodcasts searches for podcasts based on the provided parameters
func SearchPodcasts(req PodcastSearchRequest) (*PodcastSearchResponse, error) {
	logger.Logger.Debug().
		Str("query", req.Query).
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Msg("Searching podcasts")

	// Set default limit if not specified
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Cap maximum limit
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Build database search parameters
	params := database.SearchPodcastsParams{
		Query:  req.Query,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	// Execute search
	podcasts, totalCount, err := database.SearchPodcasts(params)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("query", req.Query).
			Msg("Failed to search podcasts")
		return nil, err
	}

	// Calculate if there are more results
	hasMore := int64(req.Offset+req.Limit) < totalCount

	response := &PodcastSearchResponse{
		Podcasts:   podcasts,
		TotalCount: totalCount,
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    hasMore,
	}

	logger.Logger.Debug().
		Int64("total_count", totalCount).
		Int("results", len(podcasts)).
		Bool("has_more", hasMore).
		Msg("Podcast search completed")

	return response, nil
}

// GetEpisodeById retrieves a single episode by its YouTube video ID
func GetEpisodeById(youtubeVideoId string) (*models.PodcastEpisode, error) {
	logger.Logger.Debug().
		Str("youtube_video_id", youtubeVideoId).
		Msg("Getting episode by ID")

	episode, err := database.GetEpisodeByYoutubeVideoId(youtubeVideoId)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("youtube_video_id", youtubeVideoId).
			Msg("Failed to get episode")
		return nil, err
	}

	return episode, nil
}
