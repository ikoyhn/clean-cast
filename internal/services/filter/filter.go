package filter

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"regexp"
	"strings"
	"time"
)

// FilterService handles content filtering logic
type FilterService struct {
	filters []models.Filter
}

// NewFilterService creates a new filter service with the given filters
func NewFilterService(filters []models.Filter) *FilterService {
	return &FilterService{
		filters: filters,
	}
}

// FilterEpisode determines if an episode should be blocked based on filters
func (fs *FilterService) FilterEpisode(episode *models.PodcastEpisode) bool {
	// If no filters, allow the episode
	if len(fs.filters) == 0 {
		return true
	}

	// Track if any "allow" filter matches
	allowMatched := false
	// Track if any "block" filter matches
	blockMatched := false

	for _, filter := range fs.filters {
		if !filter.Enabled {
			continue
		}

		// Check if filter applies to this feed
		if filter.FeedId != nil && *filter.FeedId != episode.PodcastId {
			continue
		}

		matched := false

		switch filter.FilterType {
		case "keyword":
			matched = fs.matchKeyword(episode, filter)
		case "regex":
			matched = fs.matchRegex(episode, filter)
		case "video_id":
			matched = fs.matchVideoId(episode, filter)
		case "duration":
			matched = fs.matchDuration(episode, filter)
		}

		if matched {
			if filter.Action == "allow" {
				allowMatched = true
			} else if filter.Action == "block" {
				blockMatched = true
				logger.Logger.Debug().
					Str("video_id", episode.YoutubeVideoId).
					Str("filter_name", filter.Name).
					Msg("Episode blocked by filter")
			}
		}
	}

	// If explicitly allowed, allow it
	if allowMatched {
		return true
	}

	// If blocked and not allowed, block it
	if blockMatched {
		return false
	}

	// Default: allow
	return true
}

// FilterEpisodes filters a slice of episodes
func (fs *FilterService) FilterEpisodes(episodes []models.PodcastEpisode) []models.PodcastEpisode {
	var filtered []models.PodcastEpisode
	for _, episode := range episodes {
		if fs.FilterEpisode(&episode) {
			filtered = append(filtered, episode)
		}
	}
	return filtered
}

// matchKeyword checks if keyword filter matches
func (fs *FilterService) matchKeyword(episode *models.PodcastEpisode, filter models.Filter) bool {
	pattern := filter.Pattern
	if !filter.CaseSensitive {
		pattern = strings.ToLower(pattern)
	}

	if filter.ApplyToTitle {
		title := episode.EpisodeName
		if !filter.CaseSensitive {
			title = strings.ToLower(title)
		}
		if strings.Contains(title, pattern) {
			return true
		}
	}

	if filter.ApplyToDescription {
		description := episode.EpisodeDescription
		if !filter.CaseSensitive {
			description = strings.ToLower(description)
		}
		if strings.Contains(description, pattern) {
			return true
		}
	}

	return false
}

// matchRegex checks if regex filter matches
func (fs *FilterService) matchRegex(episode *models.PodcastEpisode, filter models.Filter) bool {
	var pattern string
	if filter.CaseSensitive {
		pattern = filter.Pattern
	} else {
		pattern = "(?i)" + filter.Pattern
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		logger.Logger.Error().Err(err).Str("pattern", filter.Pattern).Msg("Invalid regex pattern")
		return false
	}

	if filter.ApplyToTitle && regex.MatchString(episode.EpisodeName) {
		return true
	}

	if filter.ApplyToDescription && regex.MatchString(episode.EpisodeDescription) {
		return true
	}

	return false
}

// matchVideoId checks if video ID filter matches
func (fs *FilterService) matchVideoId(episode *models.PodcastEpisode, filter models.Filter) bool {
	return episode.YoutubeVideoId == filter.Pattern
}

// matchDuration checks if duration filter matches
func (fs *FilterService) matchDuration(episode *models.PodcastEpisode, filter models.Filter) bool {
	durationSeconds := int(episode.Duration.Seconds())

	if filter.MinDuration != nil && durationSeconds < *filter.MinDuration {
		return true
	}

	if filter.MaxDuration != nil && durationSeconds > *filter.MaxDuration {
		return true
	}

	return false
}

// ApplyFeedPreferences applies feed-specific preferences to filter episodes
func ApplyFeedPreferences(episodes []models.PodcastEpisode, feedPrefs *models.FeedPreferences) []models.PodcastEpisode {
	if feedPrefs == nil {
		return episodes
	}

	var filtered []models.PodcastEpisode
	for _, episode := range episodes {
		// Apply duration filters
		if feedPrefs.MinDuration != nil {
			minDur := time.Duration(*feedPrefs.MinDuration) * time.Second
			if episode.Duration < minDur {
				continue
			}
		}

		if feedPrefs.MaxDuration != nil {
			maxDur := time.Duration(*feedPrefs.MaxDuration) * time.Second
			if episode.Duration > maxDur {
				continue
			}
		}

		filtered = append(filtered, episode)
	}

	return filtered
}

// ValidateFilter validates a filter before creation/update
func ValidateFilter(filter *models.Filter) error {
	// Validate filter type
	validTypes := map[string]bool{
		"keyword":  true,
		"regex":    true,
		"video_id": true,
		"duration": true,
	}
	if !validTypes[filter.FilterType] {
		return &FilterError{Message: "Invalid filter type"}
	}

	// Validate action
	validActions := map[string]bool{
		"block": true,
		"allow": true,
	}
	if !validActions[filter.Action] {
		return &FilterError{Message: "Invalid action"}
	}

	// Validate pattern for regex filters
	if filter.FilterType == "regex" {
		_, err := regexp.Compile(filter.Pattern)
		if err != nil {
			return &FilterError{Message: "Invalid regex pattern", Cause: err}
		}
	}

	// Validate pattern is not empty for non-duration filters
	if filter.FilterType != "duration" && filter.Pattern == "" {
		return &FilterError{Message: "Pattern cannot be empty"}
	}

	// Validate duration filters have at least one duration bound
	if filter.FilterType == "duration" {
		if filter.MinDuration == nil && filter.MaxDuration == nil {
			return &FilterError{Message: "Duration filter must specify min_duration or max_duration"}
		}
	}

	return nil
}

// FilterError represents a filter validation error
type FilterError struct {
	Message string
	Cause   error
}

func (e *FilterError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}
