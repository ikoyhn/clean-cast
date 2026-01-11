package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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

func UpdatePodcast(podcast *models.Podcast) {
	db.Save(podcast)
}
