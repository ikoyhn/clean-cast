package database

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"time"

	"gorm.io/gorm"
)

// GetDownloadProgress retrieves the download progress for a video
func GetDownloadProgress(videoId string) (*models.DownloadProgress, error) {
	var progress models.DownloadProgress
	result := db.Where("video_id = ?", videoId).First(&progress)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		logger.Logger.Error().
			Err(result.Error).
			Str("video_id", videoId).
			Msg("Failed to get download progress")
		return nil, result.Error
	}

	return &progress, nil
}

// CreateDownloadProgress creates a new download progress entry
func CreateDownloadProgress(progress *models.DownloadProgress) error {
	result := db.Create(progress)
	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Str("video_id", progress.VideoId).
			Msg("Failed to create download progress")
		return result.Error
	}

	logger.Logger.Debug().
		Str("video_id", progress.VideoId).
		Str("status", string(progress.Status)).
		Msg("Created download progress entry")

	return nil
}

// UpdateDownloadProgress updates an existing download progress entry
func UpdateDownloadProgress(progress *models.DownloadProgress) error {
	result := db.Save(progress)
	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Str("video_id", progress.VideoId).
			Msg("Failed to update download progress")
		return result.Error
	}

	logger.Logger.Debug().
		Str("video_id", progress.VideoId).
		Str("status", string(progress.Status)).
		Float64("progress", progress.ProgressPercent).
		Msg("Updated download progress")

	return nil
}

// DeleteDownloadProgress deletes a download progress entry
func DeleteDownloadProgress(videoId string) error {
	result := db.Where("video_id = ?", videoId).Delete(&models.DownloadProgress{})
	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Str("video_id", videoId).
			Msg("Failed to delete download progress")
		return result.Error
	}

	logger.Logger.Debug().
		Str("video_id", videoId).
		Msg("Deleted download progress entry")

	return nil
}

// GetOrCreateDownloadProgress gets existing progress or creates a new one
func GetOrCreateDownloadProgress(videoId, audioFormat string) (*models.DownloadProgress, error) {
	progress, err := GetDownloadProgress(videoId)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		// Create new progress entry
		progress = &models.DownloadProgress{
			VideoId:     videoId,
			Status:      models.DownloadStatusPending,
			AudioFormat: audioFormat,
		}
		err = CreateDownloadProgress(progress)
		if err != nil {
			return nil, err
		}
	}

	return progress, nil
}

// MarkDownloadInProgress marks a download as in progress
func MarkDownloadInProgress(videoId, tempFilePath string) error {
	progress, err := GetDownloadProgress(videoId)
	if err != nil {
		return err
	}

	if progress == nil {
		logger.Logger.Warn().
			Str("video_id", videoId).
			Msg("Download progress not found, cannot mark as in progress")
		return gorm.ErrRecordNotFound
	}

	progress.MarkAsInProgress()
	progress.TempFilePath = tempFilePath

	return UpdateDownloadProgress(progress)
}

// MarkDownloadCompleted marks a download as completed
func MarkDownloadCompleted(videoId, finalFilePath string) error {
	progress, err := GetDownloadProgress(videoId)
	if err != nil {
		return err
	}

	if progress == nil {
		logger.Logger.Warn().
			Str("video_id", videoId).
			Msg("Download progress not found, cannot mark as completed")
		return gorm.ErrRecordNotFound
	}

	progress.MarkAsCompleted()
	progress.FinalFilePath = finalFilePath

	return UpdateDownloadProgress(progress)
}

// MarkDownloadFailed marks a download as failed with an error
func MarkDownloadFailed(videoId string, err error) error {
	progress, dbErr := GetDownloadProgress(videoId)
	if dbErr != nil {
		return dbErr
	}

	if progress == nil {
		logger.Logger.Warn().
			Str("video_id", videoId).
			Msg("Download progress not found, cannot mark as failed")
		return gorm.ErrRecordNotFound
	}

	progress.MarkAsFailed(err)
	progress.IncrementRetry()

	return UpdateDownloadProgress(progress)
}

// UpdateDownloadProgressBytes updates the bytes downloaded and total bytes
func UpdateDownloadProgressBytes(videoId string, bytesDownloaded, totalBytes int64) error {
	progress, err := GetDownloadProgress(videoId)
	if err != nil {
		return err
	}

	if progress == nil {
		return nil // Silently ignore if not found
	}

	progress.UpdateProgress(bytesDownloaded, totalBytes)

	return UpdateDownloadProgress(progress)
}

// GetIncompleteDownloads retrieves all incomplete downloads
func GetIncompleteDownloads() ([]models.DownloadProgress, error) {
	var downloads []models.DownloadProgress
	result := db.Where("status IN ?", []models.DownloadStatus{
		models.DownloadStatusPending,
		models.DownloadStatusInProgress,
		models.DownloadStatusFailed,
	}).Find(&downloads)

	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Msg("Failed to get incomplete downloads")
		return nil, result.Error
	}

	return downloads, nil
}

// GetFailedDownloads retrieves all failed downloads that can be retried
func GetFailedDownloads(maxRetries int) ([]models.DownloadProgress, error) {
	var downloads []models.DownloadProgress
	result := db.Where("status = ? AND retry_count < ?",
		models.DownloadStatusFailed, maxRetries).Find(&downloads)

	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Msg("Failed to get failed downloads")
		return nil, result.Error
	}

	return downloads, nil
}

// CleanupOldDownloads removes download progress entries older than the specified duration
func CleanupOldDownloads(olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)

	result := db.Where("status = ? AND updated_at < ?",
		models.DownloadStatusCompleted, cutoffTime).
		Delete(&models.DownloadProgress{})

	if result.Error != nil {
		logger.Logger.Error().
			Err(result.Error).
			Msg("Failed to cleanup old downloads")
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Logger.Info().
			Int64("rows_deleted", result.RowsAffected).
			Msg("Cleaned up old download progress entries")
	}

	return nil
}

// GetDownloadStats returns statistics about downloads
func GetDownloadStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	for _, status := range []models.DownloadStatus{
		models.DownloadStatusPending,
		models.DownloadStatusInProgress,
		models.DownloadStatusCompleted,
		models.DownloadStatusFailed,
		models.DownloadStatusCancelled,
	} {
		var count int64
		result := db.Model(&models.DownloadProgress{}).
			Where("status = ?", status).
			Count(&count)

		if result.Error != nil {
			logger.Logger.Error().
				Err(result.Error).
				Str("status", string(status)).
				Msg("Failed to count downloads by status")
			continue
		}

		stats[string(status)] = count
	}

	return stats, nil
}
