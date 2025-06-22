package database

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDatabase() {
	// Get the current working directory.
	dir, err1 := os.Getwd()
	if err1 != nil {
		fmt.Println("Error getting current directory:", err1)
		return
	}

	// Print the current working directory.
	fmt.Println("Current working directory:", dir)
	var err error
	// Create the database file if it doesn't exist
	if _, err := os.Stat("C:/Users/jared/Documents/code/sqlite.db"); os.IsNotExist(err) {
		err := os.MkdirAll("C:/Users/jared/Documents/code/", os.ModePerm)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("C:/Users/jared/Documents/code/sqlite.db")
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			return
		}
	}

	db, err = gorm.Open(sqlite.Open("C:/Users/jared/Documents/code/sqlite.db"), &gorm.Config{
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
