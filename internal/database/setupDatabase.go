package database

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDatabase() {
	var err error
	// Create the database file if it doesn't exist
	if _, err := os.Stat(config.Config.DbFile); os.IsNotExist(err) {
		err := os.MkdirAll(config.Config.ConfigDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		f, err := os.Create(config.Config.DbFile)
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			return
		}
	}

	db, err = gorm.Open(sqlite.Open(config.Config.DbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	// Run database migrations
	err = RunMigrations(db)
	if err != nil {
		panic(err)
	}
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// CheckHealth performs a health check on the database
func CheckHealth() error {
	if db == nil {
		return gorm.ErrInvalidDB
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// Close closes the database connection
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
