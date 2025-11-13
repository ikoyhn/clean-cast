package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// ===== User Preferences =====

// GetOrCreateUserPreferences gets user preferences or creates default ones
func GetOrCreateUserPreferences() (*models.UserPreferences, error) {
	var prefs models.UserPreferences
	err := db.First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default preferences
			prefs = models.UserPreferences{
				DefaultQuality:    "best",
				DefaultCategories: "[]",
				AutoDownload:      false,
			}
			err = db.Create(&prefs).Error
			if err != nil {
				return nil, err
			}
			return &prefs, nil
		}
		return nil, err
	}
	return &prefs, nil
}

// UpdateUserPreferences updates user preferences
func UpdateUserPreferences(prefs *models.UserPreferences) error {
	return db.Save(prefs).Error
}

// ===== Feed Preferences =====

// GetFeedPreferences gets preferences for a specific feed
func GetFeedPreferences(feedId string) (*models.FeedPreferences, error) {
	var prefs models.FeedPreferences
	err := db.Where("feed_id = ?", feedId).First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No preferences for this feed
		}
		return nil, err
	}
	return &prefs, nil
}

// CreateOrUpdateFeedPreferences creates or updates feed preferences
func CreateOrUpdateFeedPreferences(prefs *models.FeedPreferences) error {
	var existing models.FeedPreferences
	err := db.Where("feed_id = ?", prefs.FeedId).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new
			return db.Create(prefs).Error
		}
		return err
	}

	// Update existing
	prefs.Id = existing.Id
	prefs.CreatedAt = existing.CreatedAt
	return db.Save(prefs).Error
}

// DeleteFeedPreferences deletes preferences for a specific feed
func DeleteFeedPreferences(feedId string) error {
	return db.Where("feed_id = ?", feedId).Delete(&models.FeedPreferences{}).Error
}

// GetAllFeedPreferences gets all feed preferences
func GetAllFeedPreferences() ([]models.FeedPreferences, error) {
	var prefs []models.FeedPreferences
	err := db.Find(&prefs).Error
	if err != nil {
		return nil, err
	}
	return prefs, nil
}

// ===== Filters =====

// CreateFilter creates a new filter
func CreateFilter(filter *models.Filter) error {
	return db.Create(filter).Error
}

// GetFilter gets a filter by ID
func GetFilter(id int32) (*models.Filter, error) {
	var filter models.Filter
	err := db.First(&filter, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &filter, nil
}

// GetAllFilters gets all filters
func GetAllFilters() ([]models.Filter, error) {
	var filters []models.Filter
	err := db.Where("enabled = ?", true).Order("id ASC").Find(&filters).Error
	if err != nil {
		return nil, err
	}
	return filters, nil
}

// GetFiltersForFeed gets filters applicable to a specific feed
func GetFiltersForFeed(feedId string) ([]models.Filter, error) {
	var filters []models.Filter
	err := db.Where("enabled = ? AND (feed_id IS NULL OR feed_id = ?)", true, feedId).
		Order("id ASC").
		Find(&filters).Error
	if err != nil {
		return nil, err
	}
	return filters, nil
}

// UpdateFilter updates a filter
func UpdateFilter(filter *models.Filter) error {
	return db.Save(filter).Error
}

// DeleteFilter deletes a filter
func DeleteFilter(id int32) error {
	return db.Delete(&models.Filter{}, id).Error
}

// ToggleFilter enables or disables a filter
func ToggleFilter(id int32, enabled bool) error {
	return db.Model(&models.Filter{}).Where("id = ?", id).Update("enabled", enabled).Error
}
