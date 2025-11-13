package database

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

// SaveSmartPlaylist creates a new smart playlist
func SaveSmartPlaylist(playlist *models.SmartPlaylist) error {
	return db.Create(playlist).Error
}

// GetSmartPlaylist retrieves a smart playlist by ID
func GetSmartPlaylist(id string) (*models.SmartPlaylist, error) {
	var playlist models.SmartPlaylist
	err := db.Where("id = ?", id).First(&playlist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &playlist, nil
}

// GetAllSmartPlaylists retrieves all smart playlists
func GetAllSmartPlaylists() ([]models.SmartPlaylist, error) {
	var playlists []models.SmartPlaylist
	err := db.Find(&playlists).Error
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

// UpdateSmartPlaylist updates an existing smart playlist
func UpdateSmartPlaylist(playlist *models.SmartPlaylist) error {
	return db.Save(playlist).Error
}

// DeleteSmartPlaylist deletes a smart playlist by ID
func DeleteSmartPlaylist(id string) error {
	return db.Where("id = ?", id).Delete(&models.SmartPlaylist{}).Error
}

// SmartPlaylistExists checks if a smart playlist exists
func SmartPlaylistExists(id string) (bool, error) {
	var playlist models.SmartPlaylist
	err := db.Where("id = ?", id).First(&playlist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetAllEpisodes retrieves all podcast episodes
func GetAllEpisodes() ([]models.PodcastEpisode, error) {
	var episodes []models.PodcastEpisode
	err := db.Order("published_date DESC").Find(&episodes).Error
	if err != nil {
		return nil, err
	}
	return episodes, nil
}
