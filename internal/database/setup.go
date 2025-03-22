package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func HistoryDatabaseConnect() {
	var err error
	// Create the database file if it doesn't exist
	if _, err := os.Stat("/config/sqlite.db"); os.IsNotExist(err) {
		err := os.MkdirAll("/config", os.ModePerm)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("/config/sqlite.db")
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	db, err = gorm.Open(sqlite.Open("/config/sqlite.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&models.EpisodePlaybackHistory{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&models.PodcastEpisode{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&models.Podcast{})
	if err != nil {
		panic(err)
	}
}
