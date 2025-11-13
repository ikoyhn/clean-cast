package models

import (
	"time"
)

// SmartPlaylist represents a dynamic playlist with filtering rules
type SmartPlaylist struct {
	Id          string    `json:"id" gorm:"primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Rules       string    `json:"rules" gorm:"type:text"` // JSON-encoded rules
	Logic       string    `json:"logic" gorm:"default:'AND'"` // AND or OR
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SmartPlaylistRule represents a single filter rule
type SmartPlaylistRule struct {
	Field    string      `json:"field"`    // duration, publish_date, keyword, title, description, channel_id
	Operator string      `json:"operator"` // equals, contains, greater_than, less_than, before, after
	Value    interface{} `json:"value"`
}

// SmartPlaylistRules represents a collection of rules
type SmartPlaylistRules struct {
	Rules []SmartPlaylistRule `json:"rules"`
}

// SmartPlaylistCreateRequest represents the request to create a smart playlist
type SmartPlaylistCreateRequest struct {
	Name        string                `json:"name" validate:"required"`
	Description string                `json:"description"`
	Rules       []SmartPlaylistRule   `json:"rules" validate:"required,min=1"`
	Logic       string                `json:"logic" validate:"oneof=AND OR"`
}

// SmartPlaylistUpdateRequest represents the request to update a smart playlist
type SmartPlaylistUpdateRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Rules       []SmartPlaylistRule   `json:"rules"`
	Logic       string                `json:"logic" validate:"oneof=AND OR"`
}

// SmartPlaylistResponse represents the response for a smart playlist
type SmartPlaylistResponse struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Rules       []SmartPlaylistRule   `json:"rules"`
	Logic       string                `json:"logic"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}
