package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Webhook Config operations

func GetWebhookConfig(id int32) *models.WebhookConfig {
	var config models.WebhookConfig
	err := db.Where("id = ?", id).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return nil
	}
	return &config
}

func GetAllWebhookConfigs() []models.WebhookConfig {
	var configs []models.WebhookConfig
	db.Find(&configs)
	return configs
}

func GetEnabledWebhookConfigs() []models.WebhookConfig {
	var configs []models.WebhookConfig
	db.Where("enabled = ?", true).Find(&configs)
	return configs
}

func GetWebhookConfigsByEvent(event string) []models.WebhookConfig {
	var configs []models.WebhookConfig
	// Find configs where the event is in the comma-separated events list
	db.Where("enabled = ? AND (events LIKE ? OR events LIKE ? OR events LIKE ? OR events = ?)",
		true,
		event+",%",
		"%,"+event+",%",
		"%,"+event,
		event,
	).Find(&configs)
	return configs
}

func SaveWebhookConfig(config *models.WebhookConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	return db.Create(config).Error
}

func UpdateWebhookConfig(config *models.WebhookConfig) error {
	config.UpdatedAt = time.Now()
	return db.Save(config).Error
}

func DeleteWebhookConfig(id int32) error {
	return db.Delete(&models.WebhookConfig{}, id).Error
}

// Webhook Delivery operations

func GetWebhookDelivery(id int32) *models.WebhookDelivery {
	var delivery models.WebhookDelivery
	err := db.Where("id = ?", id).First(&delivery).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return nil
	}
	return &delivery
}

func GetWebhookDeliveriesByConfig(configId int32, limit int) []models.WebhookDelivery {
	var deliveries []models.WebhookDelivery
	db.Where("webhook_config_id = ?", configId).
		Order("created_at DESC").
		Limit(limit).
		Find(&deliveries)
	return deliveries
}

func GetPendingWebhookRetries() []models.WebhookDelivery {
	var deliveries []models.WebhookDelivery
	now := time.Now()
	db.Where("status = ? AND next_retry_at <= ? AND attempts < max_retries",
		"retrying", now).
		Find(&deliveries)
	return deliveries
}

func SaveWebhookDelivery(delivery *models.WebhookDelivery) error {
	delivery.CreatedAt = time.Now()
	delivery.UpdatedAt = time.Now()
	return db.Create(delivery).Error
}

func UpdateWebhookDelivery(delivery *models.WebhookDelivery) error {
	delivery.UpdatedAt = time.Now()
	return db.Save(delivery).Error
}

func DeleteOldWebhookDeliveries(olderThan time.Time) error {
	return db.Where("created_at < ?", olderThan).Delete(&models.WebhookDelivery{}).Error
}

// Batch Job operations

func GetBatchJobStatus(id int32) *models.BatchJobStatus {
	var job models.BatchJobStatus
	err := db.Where("id = ?", id).First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return nil
	}
	return &job
}

func GetBatchJobsByType(jobType string, limit int) []models.BatchJobStatus {
	var jobs []models.BatchJobStatus
	db.Where("job_type = ?", jobType).
		Order("created_at DESC").
		Limit(limit).
		Find(&jobs)
	return jobs
}

func GetRecentBatchJobs(limit int) []models.BatchJobStatus {
	var jobs []models.BatchJobStatus
	db.Order("created_at DESC").
		Limit(limit).
		Find(&jobs)
	return jobs
}

func SaveBatchJob(job *models.BatchJobStatus) error {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	return db.Create(job).Error
}

func UpdateBatchJob(job *models.BatchJobStatus) error {
	job.UpdatedAt = time.Now()
	return db.Save(job).Error
}

func CompleteBatchJob(jobId int32, status string, errorMessage string) error {
	now := time.Now()
	return db.Model(&models.BatchJobStatus{}).
		Where("id = ?", jobId).
		Updates(map[string]interface{}{
			"status":        status,
			"error_message": errorMessage,
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
}

func UpdateBatchJobProgress(jobId int32, processed int, failed int) error {
	return db.Model(&models.BatchJobStatus{}).
		Where("id = ?", jobId).
		Updates(map[string]interface{}{
			"processed_items": processed,
			"failed_items":    failed,
			"updated_at":      time.Now(),
		}).Error
}

// Helper function to check if webhook config supports an event
func WebhookSupportsEvent(config *models.WebhookConfig, event string) bool {
	events := strings.Split(config.Events, ",")
	for _, e := range events {
		if strings.TrimSpace(e) == event {
			return true
		}
	}
	return false
}
