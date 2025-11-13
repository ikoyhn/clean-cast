package database

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TrackPlay increments play count and updates last played time
func TrackPlay(episodeId string, ipAddress string, country string) error {
	logger.Logger.Debug().
		Str("episode_id", episodeId).
		Str("ip", ipAddress).
		Str("country", country).
		Msg("Tracking episode play")

	// First, try to find existing analytics record
	var analytics models.Analytics
	result := db.Where("episode_id = ?", episodeId).First(&analytics)

	if result.Error != nil {
		// Create new analytics record
		analytics = models.Analytics{
			EpisodeId:       episodeId,
			PlayCount:       1,
			TotalListenTime: 0,
			LastPlayed:      time.Now(),
			IPAddress:       ipAddress,
			Country:         country,
		}
		return db.Create(&analytics).Error
	}

	// Update existing record
	analytics.PlayCount++
	analytics.LastPlayed = time.Now()
	analytics.IPAddress = ipAddress
	analytics.Country = country

	return db.Save(&analytics).Error
}

// TrackListenTime updates the total listen time for an episode
func TrackListenTime(episodeId string, listenTime float64) error {
	logger.Logger.Debug().
		Str("episode_id", episodeId).
		Float64("listen_time", listenTime).
		Msg("Tracking listen time")

	return db.Model(&models.Analytics{}).
		Where("episode_id = ?", episodeId).
		UpdateColumn("total_listen_time", gorm.Expr("total_listen_time + ?", listenTime)).
		Error
}

// GetEpisodeAnalytics retrieves analytics for a specific episode
func GetEpisodeAnalytics(episodeId string) (*models.Analytics, error) {
	var analytics models.Analytics
	err := db.Where("episode_id = ?", episodeId).First(&analytics).Error
	if err != nil {
		return nil, err
	}
	return &analytics, nil
}

// GetPopularEpisodes retrieves the most popular episodes by play count
func GetPopularEpisodes(limit int, days int) ([]models.Analytics, error) {
	var analytics []models.Analytics

	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := db.Where("last_played >= ?", cutoffDate).
		Order("play_count DESC").
		Limit(limit).
		Find(&analytics).Error

	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// GetAnalyticsSummary retrieves overall analytics summary
func GetAnalyticsSummary() (map[string]interface{}, error) {
	var totalPlays int64
	var totalListenTime float64
	var uniqueEpisodes int64

	// Count total plays
	db.Model(&models.Analytics{}).Select("SUM(play_count)").Scan(&totalPlays)

	// Sum total listen time
	db.Model(&models.Analytics{}).Select("SUM(total_listen_time)").Scan(&totalListenTime)

	// Count unique episodes
	db.Model(&models.Analytics{}).Count(&uniqueEpisodes)

	summary := map[string]interface{}{
		"total_plays":       totalPlays,
		"total_listen_time": totalListenTime,
		"unique_episodes":   uniqueEpisodes,
		"avg_listen_time":   0.0,
	}

	if uniqueEpisodes > 0 {
		summary["avg_listen_time"] = totalListenTime / float64(uniqueEpisodes)
	}

	return summary, nil
}

// GetGeographicDistribution retrieves analytics grouped by country
func GetGeographicDistribution(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := db.Model(&models.Analytics{}).
		Select("country, SUM(play_count) as total_plays, COUNT(DISTINCT episode_id) as unique_episodes").
		Where("country != ''").
		Group("country").
		Order("total_plays DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// UpsertAnalytics creates or updates analytics record
func UpsertAnalytics(analytics *models.Analytics) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "episode_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"play_count", "total_listen_time", "last_played", "ip_address", "country"}),
	}).Create(analytics).Error
}
