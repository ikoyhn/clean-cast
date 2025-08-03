package database

import (
	"errors"
	"ikoyhn/podcast-sponsorblock/internal/common"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"
	"time"

	log "github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

func UpdateEpisodePlaybackHistory(youtubeVideoId string, totalTimeSkipped float64) {
	log.Info("[DB] Updating episode playback history...")
	db.Model(&models.EpisodePlaybackHistory{}).
		Where("youtube_video_id = ?", youtubeVideoId).
		FirstOrCreate(&models.EpisodePlaybackHistory{
			YoutubeVideoId:   youtubeVideoId,
			LastAccessDate:   time.Now().Unix(),
			TotalTimeSkipped: totalTimeSkipped,
		})
}

func DeleteEpisodePlaybackHistory(youtubeVideoId string) {
	db.Where("youtube_video_id = ?", youtubeVideoId).Delete(&models.EpisodePlaybackHistory{})
}

func DeletePodcastCronJob() {
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

	var histories []models.EpisodePlaybackHistory
	db.Where("last_access_date < ?", oneWeekAgo).Find(&histories)

	for _, history := range histories {
		os.Remove(config.Config.AudioDir + history.YoutubeVideoId + ".m4a")
		db.Delete(&history)
		log.Info("[DB] Deleted old episode playback history... " + history.YoutubeVideoId)
	}
}

func TrackEpisodeFiles() {
	log.Info("[DB] Tracking existing episode files...")

	if _, err := os.Stat(config.Config.AudioDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.AudioDir, 0755)
	}
	if _, err := os.Stat(config.Config.ConfigDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.ConfigDir, 0755)
	}
	files, err := os.ReadDir(config.Config.AudioDir)
	if err != nil {
		log.Fatal(err)
	}

	dbFiles := make([]string, 0)
	db.Model(&models.EpisodePlaybackHistory{}).Pluck("YoutubeVideoId", &dbFiles)

	missingFiles := make([]string, 0)
	nonExistentDbFiles := make([]string, 0)
	for _, file := range files {
		filename := file.Name()
		if !common.IsValidFilename(filename) {
			continue
		}
		found := false
		for _, dbFile := range dbFiles {
			if dbFile == filename[:len(filename)-4] {
				found = true
				break
			}
		}
		if !found {
			missingFiles = append(missingFiles, filename)
		}
	}

	for _, dbFile := range dbFiles {
		found := false
		for _, file := range files {
			if dbFile == file.Name()[:len(file.Name())-4] {
				found = true
				break
			}
		}
		if !found {
			nonExistentDbFiles = append(nonExistentDbFiles, dbFile)
		}
	}

	for _, filename := range missingFiles {
		id := filename[:len(filename)-4]
		if !common.IsValidID(id) {
			continue
		}
		db.Create(&models.EpisodePlaybackHistory{YoutubeVideoId: id, LastAccessDate: time.Now().Unix(), TotalTimeSkipped: 0})
	}

	for _, dbFile := range nonExistentDbFiles {
		if !common.IsValidID(dbFile) {
			continue
		}
		db.Where("youtube_video_id = ?", dbFile).Delete(&models.EpisodePlaybackHistory{})
		log.Info("[DB] Deleted non-existent episode playback history... " + dbFile)
	}
}

func GetEpisodePlaybackHistory(youtubeVideoId string) *models.EpisodePlaybackHistory {
	var history models.EpisodePlaybackHistory
	db.Where("youtube_video_id = ?", youtubeVideoId).First(&history)
	return &history
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

func PodcastExists(podcastId string) (bool, error) {
	var episode models.Podcast
	err := db.Where("id = ?", podcastId).First(&episode).Error
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &episode, nil
}

func GetPodcastEpisodesByPodcastId(podcastId string) ([]models.PodcastEpisode, error) {
	var episodes []models.PodcastEpisode
	err := db.Where("podcast_id = ?", podcastId).Find(&episodes).Error
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func SavePlaylistEpisodes(playlistEpisodes []models.PodcastEpisode) {
	db.CreateInBatches(playlistEpisodes, 100)
}

func GetPodcast(id string) *models.Podcast {
	var podcastDb models.Podcast
	err := db.Where("id = ?", id).Find(&podcastDb).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
	}
	if podcastDb.Id == "" {
		return nil
	}
	return &podcastDb
}

func SavePodcast(podcast *models.Podcast) {
	db.Create(&podcast)
}
