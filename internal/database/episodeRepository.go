package database

import (
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	ytApi "google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"
	"time"
)

func SavePlaylistEpisodes(playlistEpisodes []models.PodcastEpisode) {
	db.CreateInBatches(playlistEpisodes, 100)
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
		log.Error(err)
	}
	if exists {
		return true
	}
	return false
}

func GetPodcastEpisodesByPodcastId(podcastId string) ([]models.PodcastEpisode, error) {
	var episodes []models.PodcastEpisode
	err := db.Where("podcast_id = ?", podcastId).Find(&episodes).Error
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func DeletePodcastCronJob() {
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

	var histories []models.EpisodePlaybackHistory
	db.Where("last_access_date < ?", oneWeekAgo).Find(&histories)

	for _, history := range histories {
		err := os.Remove("/config/audio/" + history.YoutubeVideoId + ".m4a")
		if err != nil {
			return
		}
		db.Delete(&history)
		log.Info("[DB] Deleted old episode playback history... " + history.YoutubeVideoId)
	}
}
