package models

import (
	"time"
)

type RssRequestParams struct {
	Limit *int
	Date  *time.Time
}

// SearchEpisodeRequest represents the query parameters for episode search
type SearchEpisodeRequest struct {
	Query       string     `query:"q"`
	PodcastId   string     `query:"podcast_id"`
	StartDate   string     `query:"start_date"`   // Format: YYYY-MM-DD
	EndDate     string     `query:"end_date"`     // Format: YYYY-MM-DD
	MinDuration string     `query:"min_duration"` // Format: duration string (e.g., "30m", "1h")
	MaxDuration string     `query:"max_duration"` // Format: duration string (e.g., "30m", "1h")
	Type        string     `query:"type"`         // CHANNEL or PLAYLIST
	Limit       int        `query:"limit"`
	Offset      int        `query:"offset"`
}

// SearchPodcastRequest represents the query parameters for podcast search
type SearchPodcastRequest struct {
	Query  string `query:"q"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}
