package common

import (
	"time"

	"ikoyhn/podcast-sponsorblock/internal/logger"
)

const (
	// YouTube date format
	YouTubeDateFormat = "2006-01-02T15:04:05Z07:00"
)

// ParseYouTubeDate parses a YouTube date string into a time.Time object
// Returns zero time if parsing fails
func ParseYouTubeDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	publishedAt, err := time.Parse(YouTubeDateFormat, dateStr)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("date_string", dateStr).
			Msg("Failed to parse YouTube date")
		return time.Time{}
	}

	return publishedAt
}

// ParseYouTubeDateOrDefault parses a YouTube date string with a default fallback
func ParseYouTubeDateOrDefault(dateStr string, defaultTime time.Time) time.Time {
	if dateStr == "" {
		return defaultTime
	}

	publishedAt, err := time.Parse(YouTubeDateFormat, dateStr)
	if err != nil {
		logger.Logger.Warn().
			Err(err).
			Str("date_string", dateStr).
			Msg("Failed to parse YouTube date, using default")
		return defaultTime
	}

	return publishedAt
}

// ParseYouTubeDateSafe parses a YouTube date string and returns an error if parsing fails
func ParseYouTubeDateSafe(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	return time.Parse(YouTubeDateFormat, dateStr)
}
