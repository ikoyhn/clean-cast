package database

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDatabase() {
	var err error
	if _, err := os.Stat(config.AppConfig.Setup.DbFile); os.IsNotExist(err) {
		err := os.MkdirAll(config.AppConfig.Setup.ConfigDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(config.AppConfig.Setup.DbFile)
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			return
		}
	}

	db, err = gorm.Open(sqlite.Open(config.AppConfig.Setup.DbFile), &gorm.Config{
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
