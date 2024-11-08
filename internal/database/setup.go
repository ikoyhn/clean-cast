package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func ConnectDatabase() {
	var err error
	db, err = gorm.Open(sqlite.Open("/config/sqlite.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&models.EpisodePlaybackHistory{})
	if err != nil {
		panic(err)
	}
}
