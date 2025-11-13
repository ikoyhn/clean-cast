package database

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"os"
	"time"

	"gorm.io/gorm/clause"
)

func UpdateEpisodePlaybackHistory(youtubeVideoId string, totalTimeSkipped float64) {
	logger.Logger.Debug().
		Str("video_id", youtubeVideoId).
		Float64("time_skipped", totalTimeSkipped).
		Msg("Updating episode playback history")
	// Use proper upsert with OnConflict to avoid race conditions
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "youtube_video_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_access_date", "total_time_skipped"}),
	}).Create(&models.EpisodePlaybackHistory{
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
	logger.Logger.Info().Msg("App started, tracking existing episode files")
	if _, err := os.Stat(config.Config.AudioDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.AudioDir, 0755)
	}
	if _, err := os.Stat(config.Config.ConfigDir); os.IsNotExist(err) {
		os.MkdirAll(config.Config.ConfigDir, 0755)
	}
	files, err := os.ReadDir(config.Config.AudioDir)
	if err != nil {
		logger.Logger.Error().Err(err).Str("dir", config.Config.AudioDir).Msg("Failed to read audio directory")
	}

	dbFiles := make([]string, 0)
	db.Model(&models.EpisodePlaybackHistory{}).Pluck("YoutubeVideoId", &dbFiles)

	// Create a map of database video IDs for O(1) lookup - Optimized from O(n²) to O(n)
	dbFileMap := make(map[string]bool, len(dbFiles))
	for _, dbFile := range dbFiles {
		dbFileMap[dbFile] = true
	}

	// Create a map of filesystem video IDs for O(1) lookup - Optimized from O(n²) to O(n)
	fileMap := make(map[string]bool)
	missingFiles := make([]string, 0)
	for _, file := range files {
		filename := file.Name()
		if !common.IsValidFilename(filename) {
			continue
		}
		videoId := filename[:len(filename)-4]
		fileMap[videoId] = true

		// Check if this file is missing from the database
		if !dbFileMap[videoId] {
			missingFiles = append(missingFiles, filename)
		}
	}

	// Find DB entries without corresponding files
	nonExistentDbFiles := make([]string, 0)
	for _, dbFile := range dbFiles {
		if !fileMap[dbFile] {
			nonExistentDbFiles = append(nonExistentDbFiles, dbFile)
		}
	}

	// Add missing files to database
	for _, filename := range missingFiles {
		id := filename[:len(filename)-4]
		if !common.IsValidID(id) {
			continue
		}
		db.Create(&models.EpisodePlaybackHistory{YoutubeVideoId: id, LastAccessDate: time.Now().Unix(), TotalTimeSkipped: 0})
	}

	// Delete non-existent files from database
	for _, dbFile := range nonExistentDbFiles {
		if !common.IsValidID(dbFile) {
			continue
		}
		db.Where("youtube_video_id = ?", dbFile).Delete(&models.EpisodePlaybackHistory{})
		logger.Logger.Info().
			Str("video_id", dbFile).
			Msg("Deleted non-existent episode playback history")
	}
}
