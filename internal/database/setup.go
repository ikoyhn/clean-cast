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
	if _, err := os.Stat("C:/Users/jared/Documents/code/config/sqlite1.db"); os.IsNotExist(err) {
		err := os.MkdirAll("C:/Users/jared/Documents/code/config", os.ModePerm)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("C:/Users/jared/Documents/code/config/sqlite1.db")
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	db, err = gorm.Open(sqlite.Open("C:/Users/jared/Documents/code/config/sqlite1.db"), &gorm.Config{
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
