package models

import (
	"time"
)

// DownloadStatus represents the status of a download
type DownloadStatus string

const (
	DownloadStatusPending    DownloadStatus = "pending"
	DownloadStatusInProgress DownloadStatus = "in_progress"
	DownloadStatusCompleted  DownloadStatus = "completed"
	DownloadStatusFailed     DownloadStatus = "failed"
	DownloadStatusCancelled  DownloadStatus = "cancelled"
)

// DownloadProgress tracks the progress of a video download
type DownloadProgress struct {
	Id               int32          `gorm:"autoIncrement;primary_key;not null"`
	VideoId          string         `json:"video_id" gorm:"uniqueIndex;not null"`
	Status           DownloadStatus `json:"status" gorm:"index;not null;default:'pending'"`
	ProgressPercent  float64        `json:"progress_percent" gorm:"default:0"`
	BytesDownloaded  int64          `json:"bytes_downloaded" gorm:"default:0"`
	TotalBytes       int64          `json:"total_bytes" gorm:"default:0"`
	RetryCount       int            `json:"retry_count" gorm:"default:0"`
	LastError        string         `json:"last_error" gorm:"type:text"`
	TempFilePath     string         `json:"temp_file_path" gorm:"type:text"`
	FinalFilePath    string         `json:"final_file_path" gorm:"type:text"`
	AudioFormat      string         `json:"audio_format" gorm:"not null;default:'m4a'"`
	StartedAt        *time.Time     `json:"started_at" gorm:"index"`
	CompletedAt      *time.Time     `json:"completed_at"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// IsComplete returns true if the download is completed
func (d *DownloadProgress) IsComplete() bool {
	return d.Status == DownloadStatusCompleted
}

// IsFailed returns true if the download has failed
func (d *DownloadProgress) IsFailed() bool {
	return d.Status == DownloadStatusFailed
}

// IsInProgress returns true if the download is in progress
func (d *DownloadProgress) IsInProgress() bool {
	return d.Status == DownloadStatusInProgress
}

// CanRetry returns true if the download can be retried
func (d *DownloadProgress) CanRetry(maxRetries int) bool {
	return d.RetryCount < maxRetries && (d.IsFailed() || d.Status == DownloadStatusPending)
}

// IncrementRetry increments the retry counter
func (d *DownloadProgress) IncrementRetry() {
	d.RetryCount++
}

// MarkAsInProgress marks the download as in progress
func (d *DownloadProgress) MarkAsInProgress() {
	d.Status = DownloadStatusInProgress
	now := time.Now()
	if d.StartedAt == nil {
		d.StartedAt = &now
	}
}

// MarkAsCompleted marks the download as completed
func (d *DownloadProgress) MarkAsCompleted() {
	d.Status = DownloadStatusCompleted
	now := time.Now()
	d.CompletedAt = &now
	d.ProgressPercent = 100.0
}

// MarkAsFailed marks the download as failed with an error message
func (d *DownloadProgress) MarkAsFailed(err error) {
	d.Status = DownloadStatusFailed
	if err != nil {
		d.LastError = err.Error()
	}
}

// UpdateProgress updates the download progress
func (d *DownloadProgress) UpdateProgress(bytesDownloaded, totalBytes int64) {
	d.BytesDownloaded = bytesDownloaded
	d.TotalBytes = totalBytes

	if totalBytes > 0 {
		d.ProgressPercent = (float64(bytesDownloaded) / float64(totalBytes)) * 100.0
	}
}
