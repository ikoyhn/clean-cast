package database

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/filter"
	"os"
	"time"

	"github.com/pkg/errors"
	ytApi "google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
)

func SavePlaylistEpisodes(playlistEpisodes []models.PodcastEpisode) {
	db.CreateInBatches(playlistEpisodes, constants.BatchSize)
}

func EpisodeExists(youtubeVideoId string, episodeType string) (bool, error) {
	var episode models.PodcastEpisode
	err := db.Where("youtube_video_id = ? AND type = ?", youtubeVideoId, episodeType).First(&episode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetLatestEpisode(podcastId string) (*models.PodcastEpisode, error) {
	var episode models.PodcastEpisode
	err := db.Where("podcast_id = ?", podcastId).Order("published_date DESC").First(&episode).Error
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

func GetOldestEpisode(podcastId string) (*models.PodcastEpisode, error) {
	var episode models.PodcastEpisode
	err := db.Where("podcast_id = ?", podcastId).Order("published_date ASC").First(&episode).Error
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

func GetAllPodcastEpisodeIds(podcastId string) ([]string, error) {
	var episodes []models.PodcastEpisode

	err := db.Where("podcast_id = ?", podcastId).Find(&episodes).Error
	if err != nil {
		return nil, err
	}

	var episodeIds []string
	for _, episode := range episodes {
		episodeIds = append(episodeIds, episode.YoutubeVideoId)
	}

	return episodeIds, nil
}

func IsEpisodeSaved(item *ytApi.Video) bool {
	exists, err := EpisodeExists(item.Id, "CHANNEL")
	if err != nil {
		logger.Logger.Error().Err(err).Str("video_id", item.Id).Msg("Failed to check if episode exists")
	}
	if exists {
		return true
	}
	return false
}

func GetPodcastEpisodesByPodcastId(podcastId string, podcastType enum.PodcastType) ([]models.PodcastEpisode, error) {
	var episodes []models.PodcastEpisode
	if podcastType == enum.PLAYLIST {
		err := db.Where("podcast_id = ?", podcastId).
			Order("published_date DESC").
			Find(&episodes).Error
		if err != nil {
			return nil, err
		}
	} else if podcastType == enum.CHANNEL {
		dur, err := time.ParseDuration(config.Config.MinDuration)
		if err != nil {
			return nil, err
		}

		err = db.Where("podcast_id = ? AND duration >= ?", podcastId, dur).
			Order("published_date DESC").
			Find(&episodes).Error
		if err != nil {
			return nil, err
		}
	}

	return episodes, nil
}

func DeletePodcastCronJob() {
	oneWeekAgo := time.Now().Add(-constants.CleanupDays * 24 * time.Hour).Unix()

	var histories []models.EpisodePlaybackHistory
	db.Where("last_access_date < ?", oneWeekAgo).Find(&histories)

	for _, history := range histories {
		err := os.Remove(config.Config.AudioDir + history.YoutubeVideoId + ".m4a")
		if err != nil {
			logger.Logger.Error().
				Err(err).
				Str("video_id", history.YoutubeVideoId).
				Msg("Failed to delete file for episode")
			continue
		}
		db.Delete(&history)
		logger.Logger.Info().
			Str("video_id", history.YoutubeVideoId).
			Msg("Deleted old episode playback history")
	}
}

// SearchEpisodesParams holds search parameters for episodes
type SearchEpisodesParams struct {
	Query           string
	PodcastId       *string
	StartDate       *time.Time
	EndDate         *time.Time
	MinDuration     *time.Duration
	MaxDuration     *time.Duration
	Type            *string
	Limit           int
	Offset          int
}

// SearchEpisodes searches for episodes based on provided parameters
func SearchEpisodes(params SearchEpisodesParams) ([]models.PodcastEpisode, int64, error) {
	var episodes []models.PodcastEpisode
	var totalCount int64

	// Build base query
	query := db.Model(&models.PodcastEpisode{})

	// Apply search query filter (search in episode name and description)
	if params.Query != "" {
		searchPattern := "%" + params.Query + "%"
		query = query.Where("episode_name LIKE ? OR episode_description LIKE ?", searchPattern, searchPattern)
	}

	// Apply podcast ID filter
	if params.PodcastId != nil && *params.PodcastId != "" {
		query = query.Where("podcast_id = ?", *params.PodcastId)
	}

	// Apply type filter (CHANNEL or PLAYLIST)
	if params.Type != nil && *params.Type != "" {
		query = query.Where("type = ?", *params.Type)
	}

	// Apply date range filters
	if params.StartDate != nil {
		query = query.Where("published_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("published_date <= ?", *params.EndDate)
	}

	// Apply duration filters
	if params.MinDuration != nil {
		query = query.Where("duration >= ?", *params.MinDuration)
	}
	if params.MaxDuration != nil {
		query = query.Where("duration <= ?", *params.MaxDuration)
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering and pagination
	query = query.Order("published_date DESC")

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Execute query
	if err := query.Find(&episodes).Error; err != nil {
		return nil, 0, err
	}

	return episodes, totalCount, nil
}

// GetEpisodeByYoutubeVideoId retrieves an episode by its YouTube video ID
func GetEpisodeByYoutubeVideoId(youtubeVideoId string) (*models.PodcastEpisode, error) {
	var episode models.PodcastEpisode
	err := db.Where("youtube_video_id = ?", youtubeVideoId).First(&episode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &episode, nil
}

// ApplyFiltersAndPreferences applies content filters and feed preferences to episodes
func ApplyFiltersAndPreferences(episodes []models.PodcastEpisode, feedId string) ([]models.PodcastEpisode, error) {
	// Get feed preferences
	feedPrefs, err := GetFeedPreferences(feedId)
	if err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to get feed preferences")
		// Don't fail if preferences can't be loaded, just continue without them
	}

	// Apply feed-specific duration filters if set
	if feedPrefs != nil {
		episodes = filter.ApplyFeedPreferences(episodes, feedPrefs)
	}

	// Get and apply content filters
	filters, err := GetFiltersForFeed(feedId)
	if err != nil {
		logger.Logger.Error().Err(err).Str("feed_id", feedId).Msg("Failed to load filters")
		// Don't fail if filters can't be loaded, just continue without them
	} else {
		filterService := filter.NewFilterService(filters)
		episodes = filterService.FilterEpisodes(episodes)
	}

	return episodes, nil
}
