package common

import (
	"time"

	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/logger"
)

// ValidateEpisodeDuration validates if an episode duration meets the minimum duration requirement
// Returns true if the episode is valid (duration > minimum)
func ValidateEpisodeDuration(duration time.Duration) bool {
	minDuration, err := GetMinDuration()
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Msg("Failed to get minimum duration, skipping validation")
		return true // Allow episode if we can't determine the minimum
	}

	return duration > minDuration
}

// GetMinDuration returns the configured minimum duration
func GetMinDuration() (time.Duration, error) {
	dur, err := time.ParseDuration(config.Config.MinDuration)
	if err != nil {
		logger.Logger.Error().
			Msgf("Invalid MIN_DURATION format '%s'. Use formats like '5m', '1h', '400s'. Error: %v",
				config.Config.MinDuration, err)
		return 0, err
	}
	return dur, nil
}

// IsEpisodeValid performs comprehensive episode validation
func IsEpisodeValid(videoId string, duration time.Duration) bool {
	if videoId == "" {
		return false
	}

	if !ValidateEpisodeDuration(duration) {
		logger.Logger.Debug().
			Str("video_id", videoId).
			Dur("duration", duration).
			Msg("Episode duration below minimum threshold")
		return false
	}

	return true
}

// ValidateVideoDuration validates video duration against minimum requirement
// and returns the duration in seconds for logging
func ValidateVideoDuration(duration time.Duration) (bool, float64) {
	seconds := duration.Seconds()
	minDur, err := GetMinDuration()
	if err != nil {
		return true, seconds // Allow if we can't determine minimum
	}

	return duration > minDur, seconds
}
