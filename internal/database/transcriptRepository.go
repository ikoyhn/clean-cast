package database

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"

	"gorm.io/gorm/clause"
)

// SaveTranscript saves or updates a transcript
func SaveTranscript(transcript *models.Transcript) error {
	logger.Logger.Debug().
		Str("episode_id", transcript.EpisodeId).
		Str("language", transcript.Language).
		Msg("Saving transcript")

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "episode_id"}, {Name: "language"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "is_auto_generated", "updated_at"}),
	}).Create(transcript).Error
}

// GetTranscript retrieves a transcript for an episode
func GetTranscript(episodeId string, language string) (*models.Transcript, error) {
	var transcript models.Transcript

	query := db.Where("episode_id = ?", episodeId)

	if language != "" {
		query = query.Where("language = ?", language)
	}

	err := query.First(&transcript).Error
	if err != nil {
		return nil, err
	}

	return &transcript, nil
}

// GetAllTranscriptsForEpisode retrieves all available transcripts for an episode
func GetAllTranscriptsForEpisode(episodeId string) ([]models.Transcript, error) {
	var transcripts []models.Transcript

	err := db.Where("episode_id = ?", episodeId).Find(&transcripts).Error
	if err != nil {
		return nil, err
	}

	return transcripts, nil
}

// TranscriptExists checks if a transcript exists for an episode
func TranscriptExists(episodeId string, language string) (bool, error) {
	var count int64

	query := db.Model(&models.Transcript{}).Where("episode_id = ?", episodeId)

	if language != "" {
		query = query.Where("language = ?", language)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteTranscript deletes a transcript
func DeleteTranscript(episodeId string, language string) error {
	logger.Logger.Debug().
		Str("episode_id", episodeId).
		Str("language", language).
		Msg("Deleting transcript")

	return db.Where("episode_id = ? AND language = ?", episodeId, language).
		Delete(&models.Transcript{}).Error
}

// GetAvailableLanguages gets all available languages for an episode's transcripts
func GetAvailableLanguages(episodeId string) ([]string, error) {
	var languages []string

	err := db.Model(&models.Transcript{}).
		Where("episode_id = ?", episodeId).
		Pluck("language", &languages).Error

	if err != nil {
		return nil, err
	}

	return languages, nil
}
