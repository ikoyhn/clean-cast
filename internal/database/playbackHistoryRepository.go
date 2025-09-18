package database

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"os"
	"time"

	"github.com/labstack/gommon/log"
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

func GetEpisodePlaybackHistory(youtubeVideoId string) *models.EpisodePlaybackHistory {
	var history models.EpisodePlaybackHistory
	db.Where("youtube_video_id = ?", youtubeVideoId).First(&history)
	return &history
}

func TrackEpisodeFiles() {
	log.Info("App started, tracking existing episode files...")
	if _, err := os.Stat(config.Config.AudioDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.AudioDir, 0755)
	}
	if _, err := os.Stat(config.Config.ConfigDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.ConfigDir, 0755)
	}
	files, err := os.ReadDir(config.Config.AudioDir)
	if err != nil {
		log.Error(err)
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
