package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// UserPreferences represents global user preferences
type UserPreferences struct {
	Id              int32     `gorm:"autoIncrement;primary_key;not null"`
	DefaultQuality  string    `json:"default_quality" gorm:"default:'best'"`    // e.g., "best", "1080p", "720p", "480p"
	DefaultCategories string  `json:"default_categories" gorm:"type:text"`      // JSON array of SponsorBlock categories
	AutoDownload    bool      `json:"auto_download" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// FeedPreferences represents per-feed preferences
type FeedPreferences struct {
	Id                    int32          `gorm:"autoIncrement;primary_key;not null"`
	FeedId                string         `json:"feed_id" gorm:"index;not null"`                    // Podcast/Channel ID
	FeedType              string         `json:"feed_type" gorm:"default:'CHANNEL'"`                // "CHANNEL" or "PLAYLIST"
	SponsorBlockCategories StringArray   `json:"sponsorblock_categories" gorm:"type:text"`         // Custom categories for this feed
	MinDuration           *int           `json:"min_duration"`                                      // Minimum duration in seconds
	MaxDuration           *int           `json:"max_duration"`                                      // Maximum duration in seconds
	Quality               string         `json:"quality" gorm:"default:'best'"`                     // Override default quality
	AutoDownload          *bool          `json:"auto_download"`                                     // Override default auto_download
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// Filter represents a content filter rule
type Filter struct {
	Id              int32      `gorm:"autoIncrement;primary_key;not null"`
	Name            string     `json:"name" gorm:"not null"`                      // User-friendly name for the filter
	FilterType      string     `json:"filter_type" gorm:"not null"`               // "keyword", "regex", "video_id", "duration"
	Action          string     `json:"action" gorm:"default:'block'"`             // "block" or "allow"
	Pattern         string     `json:"pattern" gorm:"type:text"`                  // The pattern to match (keyword, regex, video ID)
	CaseSensitive   bool       `json:"case_sensitive" gorm:"default:false"`
	ApplyToTitle    bool       `json:"apply_to_title" gorm:"default:true"`
	ApplyToDescription bool    `json:"apply_to_description" gorm:"default:false"`
	FeedId          *string    `json:"feed_id" gorm:"index"`                      // Optional: apply only to specific feed
	MinDuration     *int       `json:"min_duration"`                              // For duration filters
	MaxDuration     *int       `json:"max_duration"`                              // For duration filters
	Enabled         bool       `json:"enabled" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// StringArray is a custom type for storing string arrays in SQLite
type StringArray []string

// Scan implements the sql.Scanner interface for reading from database
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*s = []string{}
			return nil
		}
		bytes = []byte(str)
	}

	if len(bytes) == 0 {
		*s = []string{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface for writing to database
func (s StringArray) Value() (driver.Value, error) {
	if s == nil || len(s) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	return string(bytes), err
}
