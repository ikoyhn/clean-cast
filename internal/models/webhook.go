package models

import (
	"time"
)

// WebhookConfig represents a webhook configuration stored in the database
type WebhookConfig struct {
	Id        int32     `gorm:"autoIncrement;primary_key;not null"`
	Name      string    `json:"name" gorm:"not null"`
	URL       string    `json:"url" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"`   // discord, slack, generic
	Events    string    `json:"events" gorm:"not null"` // comma-separated list: new_episode, download_complete, error
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookDelivery tracks webhook delivery attempts
type WebhookDelivery struct {
	Id              int32      `gorm:"autoIncrement;primary_key;not null"`
	WebhookConfigId int32      `json:"webhook_config_id" gorm:"index;foreignkey:WebhookConfigId;association_foreignkey:Id"`
	Event           string     `json:"event" gorm:"not null"`
	Payload         string     `json:"payload" gorm:"type:text"`
	Status          string     `json:"status" gorm:"not null"` // pending, sent, failed, retrying
	ResponseCode    int        `json:"response_code"`
	ResponseBody    string     `json:"response_body" gorm:"type:text"`
	ErrorMessage    string     `json:"error_message" gorm:"type:text"`
	Attempts        int        `json:"attempts" gorm:"default:0"`
	MaxRetries      int        `json:"max_retries" gorm:"default:3"`
	NextRetryAt     *time.Time `json:"next_retry_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// WebhookEvent represents a webhook event payload
type WebhookEvent struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// BatchJobStatus represents the status of a batch operation
type BatchJobStatus struct {
	Id             int32      `gorm:"autoIncrement;primary_key;not null"`
	JobType        string     `json:"job_type" gorm:"not null;index"` // refresh, delete_episodes, add_podcasts
	Status         string     `json:"status" gorm:"not null;index"`   // pending, running, completed, failed
	TotalItems     int        `json:"total_items"`
	ProcessedItems int        `json:"processed_items"`
	FailedItems    int        `json:"failed_items"`
	ErrorMessage   string     `json:"error_message" gorm:"type:text"`
	Metadata       string     `json:"metadata" gorm:"type:text"` // JSON string with additional data
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// Batch request/response models

type BatchRefreshRequest struct {
	PodcastIds []string `json:"podcast_ids" binding:"required"`
}

type BatchDeleteEpisodesRequest struct {
	EpisodeIds []string `json:"episode_ids" binding:"required"`
}

type BatchAddPodcastsRequest struct {
	Podcasts []BatchPodcastItem `json:"podcasts" binding:"required"`
}

type BatchPodcastItem struct {
	Id   string `json:"id" binding:"required"`
	Type string `json:"type" binding:"required"` // playlist or channel
}

type BatchOperationResponse struct {
	JobId   int32  `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BatchStatusResponse struct {
	JobId          int32      `json:"job_id"`
	JobType        string     `json:"job_type"`
	Status         string     `json:"status"`
	TotalItems     int        `json:"total_items"`
	ProcessedItems int        `json:"processed_items"`
	FailedItems    int        `json:"failed_items"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}
