package database

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

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

func GetAllPodcasts() ([]models.Podcast, error) {
	var podcasts []models.Podcast
	err := db.Find(&podcasts).Error
	if err != nil {
		return nil, err
	}
	return podcasts, nil
}

// SearchPodcastsParams holds search parameters for podcasts
type SearchPodcastsParams struct {
	Query  string
	Limit  int
	Offset int
}

// SearchPodcasts searches for podcasts based on provided parameters
func SearchPodcasts(params SearchPodcastsParams) ([]models.Podcast, int64, error) {
	var podcasts []models.Podcast
	var totalCount int64

	// Build base query
	query := db.Model(&models.Podcast{})

	// Apply search query filter (search in podcast name, description, and artist name)
	if params.Query != "" {
		searchPattern := "%" + params.Query + "%"
		query = query.Where("podcast_name LIKE ? OR description LIKE ? OR artist_name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering and pagination
	query = query.Order("podcast_name ASC")

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Execute query
	if err := query.Find(&podcasts).Error; err != nil {
		return nil, 0, err
	}

	return podcasts, totalCount, nil
}

// GetAllPodcastsPaginated returns all podcasts with pagination
func GetAllPodcastsPaginated(limit, offset int) ([]models.Podcast, int64, error) {
	var podcasts []models.Podcast
	var totalCount int64

	query := db.Model(&models.Podcast{})

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	query = query.Order("podcast_name ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	if err := query.Find(&podcasts).Error; err != nil {
		return nil, 0, err
	}

	return podcasts, totalCount, nil
}
