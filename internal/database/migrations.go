package database

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
	ExecutedAt  time.Time
}

// SchemaMigration tracks executed migrations
type SchemaMigration struct {
	Version    int       `gorm:"primaryKey"`
	Name       string    `gorm:"not null"`
	ExecutedAt time.Time `gorm:"not null"`
}

// migrations holds all available migrations in order
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		Up:      migration001Up,
		Down:    migration001Down,
	},
	{
		Version: 2,
		Name:    "add_search_indexes",
		Up:      migration002Up,
		Down:    migration002Down,
	},
	{
		Version: 3,
		Name:    "add_smart_playlists",
		Up:      migration003Up,
		Down:    migration003Down,
	},
	{
		Version: 4,
		Name:    "add_preferences_and_filters",
		Up:      migration004Up,
		Down:    migration004Down,
	},
	{
		Version: 5,
		Name:    "add_analytics_and_transcripts",
		Up:      migration005Up,
		Down:    migration005Down,
	},
	{
		Version: 6,
		Name:    "add_webhooks_and_batch_jobs",
		Up:      migration006Up,
		Down:    migration006Down,
	},
	{
		Version: 7,
		Name:    "add_download_progress",
		Up:      migration007Up,
		Down:    migration007Down,
	},
}

// migration001Up creates the initial schema
func migration001Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 001: initial_schema")

	// Create Podcast table
	if err := db.AutoMigrate(&models.Podcast{}); err != nil {
		return fmt.Errorf("failed to migrate Podcast table: %w", err)
	}

	// Create PodcastEpisode table
	if err := db.AutoMigrate(&models.PodcastEpisode{}); err != nil {
		return fmt.Errorf("failed to migrate PodcastEpisode table: %w", err)
	}

	// Create EpisodePlaybackHistory table
	if err := db.AutoMigrate(&models.EpisodePlaybackHistory{}); err != nil {
		return fmt.Errorf("failed to migrate EpisodePlaybackHistory table: %w", err)
	}

	logger.Logger.Info().Msg("Migration 001 completed successfully")
	return nil
}

func migration001Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 001: initial_schema")

	if err := db.Migrator().DropTable(&models.EpisodePlaybackHistory{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.PodcastEpisode{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.Podcast{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 001 rollback completed")
	return nil
}

// migration002Up adds indexes for search optimization
func migration002Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 002: add_search_indexes")

	// Add indexes for episode search
	if !db.Migrator().HasIndex(&models.PodcastEpisode{}, "idx_episode_name") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_episode_name ON podcast_episodes(episode_name)").Error; err != nil {
			return fmt.Errorf("failed to create episode_name index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.PodcastEpisode{}, "idx_episode_description") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_episode_description ON podcast_episodes(episode_description)").Error; err != nil {
			return fmt.Errorf("failed to create episode_description index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.PodcastEpisode{}, "idx_published_date") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_published_date ON podcast_episodes(published_date)").Error; err != nil {
			return fmt.Errorf("failed to create published_date index: %w", err)
		}
	}

	// Add indexes for podcast search
	if !db.Migrator().HasIndex(&models.Podcast{}, "idx_podcast_name") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_podcast_name ON podcasts(podcast_name)").Error; err != nil {
			return fmt.Errorf("failed to create podcast_name index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.Podcast{}, "idx_podcast_description") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_podcast_description ON podcasts(description)").Error; err != nil {
			return fmt.Errorf("failed to create podcast description index: %w", err)
		}
	}

	logger.Logger.Info().Msg("Migration 002 completed successfully")
	return nil
}

func migration002Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 002: add_search_indexes")

	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_episode_name")
	db.Exec("DROP INDEX IF EXISTS idx_episode_description")
	db.Exec("DROP INDEX IF EXISTS idx_published_date")
	db.Exec("DROP INDEX IF EXISTS idx_podcast_name")
	db.Exec("DROP INDEX IF EXISTS idx_podcast_description")

	logger.Logger.Info().Msg("Migration 002 rollback completed")
	return nil
}

// migration003Up adds smart playlists table
func migration003Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 003: add_smart_playlists")

	// Create SmartPlaylist table
	if err := db.AutoMigrate(&models.SmartPlaylist{}); err != nil {
		return fmt.Errorf("failed to migrate SmartPlaylist table: %w", err)
	}

	logger.Logger.Info().Msg("Migration 003 completed successfully")
	return nil
}

func migration003Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 003: add_smart_playlists")

	if err := db.Migrator().DropTable(&models.SmartPlaylist{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 003 rollback completed")
	return nil
}

// RunMigrations executes all pending migrations
func RunMigrations(db *gorm.DB) error {
	logger.Logger.Info().Msg("Starting database migrations")

	// Create schema_migrations table if it doesn't exist
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get current version
	currentVersion := getCurrentVersion(db)
	logger.Logger.Info().Int("current_version", currentVersion).Msg("Current database version")

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			logger.Logger.Info().
				Int("version", migration.Version).
				Str("name", migration.Name).
				Msg("Running migration")

			// Start transaction
			tx := db.Begin()
			if tx.Error != nil {
				return fmt.Errorf("failed to start transaction: %w", tx.Error)
			}

			// Run migration
			if err := migration.Up(tx); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d (%s) failed: %w", migration.Version, migration.Name, err)
			}

			// Record migration
			record := SchemaMigration{
				Version:    migration.Version,
				Name:       migration.Name,
				ExecutedAt: time.Now(),
			}
			if err := tx.Create(&record).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
			}

			// Commit transaction
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
			}

			logger.Logger.Info().
				Int("version", migration.Version).
				Str("name", migration.Name).
				Msg("Migration completed successfully")
		}
	}

	finalVersion := getCurrentVersion(db)
	logger.Logger.Info().
		Int("version", finalVersion).
		Msg("All migrations completed")

	return nil
}

// getCurrentVersion returns the current migration version
func getCurrentVersion(db *gorm.DB) int {
	var migration SchemaMigration
	if err := db.Order("version DESC").First(&migration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0
		}
		logger.Logger.Error().Err(err).Msg("Failed to get current migration version")
		return 0
	}
	return migration.Version
}

// RollbackMigration rolls back the last migration
func RollbackMigration(db *gorm.DB) error {
	currentVersion := getCurrentVersion(db)
	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the migration to rollback
	var migration *Migration
	for i := range migrations {
		if migrations[i].Version == currentVersion {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration version %d not found", currentVersion)
	}

	logger.Logger.Info().
		Int("version", migration.Version).
		Str("name", migration.Name).
		Msg("Rolling back migration")

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Run rollback
	if err := migration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("rollback of migration %d failed: %w", migration.Version, err)
	}

	// Remove migration record
	if err := tx.Where("version = ?", migration.Version).Delete(&SchemaMigration{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	logger.Logger.Info().
		Int("version", migration.Version).
		Str("name", migration.Name).
		Msg("Migration rolled back successfully")

	return nil
}

// migration004Up adds user preferences, feed preferences, and content filters
func migration004Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 004: add_preferences_and_filters")

	// Create UserPreferences table
	if err := db.AutoMigrate(&models.UserPreferences{}); err != nil {
		return fmt.Errorf("failed to migrate UserPreferences table: %w", err)
	}

	// Create FeedPreferences table
	if err := db.AutoMigrate(&models.FeedPreferences{}); err != nil {
		return fmt.Errorf("failed to migrate FeedPreferences table: %w", err)
	}

	// Create Filter table
	if err := db.AutoMigrate(&models.Filter{}); err != nil {
		return fmt.Errorf("failed to migrate Filter table: %w", err)
	}

	// Add indexes for feed preferences
	if !db.Migrator().HasIndex(&models.FeedPreferences{}, "idx_feed_id") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_feed_id ON feed_preferences(feed_id)").Error; err != nil {
			return fmt.Errorf("failed to create feed_id index: %w", err)
		}
	}

	// Add indexes for filters
	if !db.Migrator().HasIndex(&models.Filter{}, "idx_filter_feed_id") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_filter_feed_id ON filters(feed_id)").Error; err != nil {
			return fmt.Errorf("failed to create filter feed_id index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.Filter{}, "idx_filter_enabled") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_filter_enabled ON filters(enabled)").Error; err != nil {
			return fmt.Errorf("failed to create filter enabled index: %w", err)
		}
	}

	logger.Logger.Info().Msg("Migration 004 completed successfully")
	return nil
}

func migration004Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 004: add_preferences_and_filters")

	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_filter_enabled")
	db.Exec("DROP INDEX IF EXISTS idx_filter_feed_id")
	db.Exec("DROP INDEX IF EXISTS idx_feed_id")

	// Drop tables
	if err := db.Migrator().DropTable(&models.Filter{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.FeedPreferences{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.UserPreferences{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 004 rollback completed")
	return nil
}

// migration005Up adds analytics and transcripts tables
func migration005Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 005: add_analytics_and_transcripts")

	// Create Analytics table
	if err := db.AutoMigrate(&models.Analytics{}); err != nil {
		return fmt.Errorf("failed to migrate Analytics table: %w", err)
	}

	// Create Transcript table
	if err := db.AutoMigrate(&models.Transcript{}); err != nil {
		return fmt.Errorf("failed to migrate Transcript table: %w", err)
	}

	// Add indexes for analytics
	if !db.Migrator().HasIndex(&models.Analytics{}, "idx_analytics_episode_id") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_analytics_episode_id ON analytics(episode_id)").Error; err != nil {
			return fmt.Errorf("failed to create analytics episode_id index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.Analytics{}, "idx_analytics_last_played") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_analytics_last_played ON analytics(last_played)").Error; err != nil {
			return fmt.Errorf("failed to create analytics last_played index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.Analytics{}, "idx_analytics_country") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_analytics_country ON analytics(country)").Error; err != nil {
			return fmt.Errorf("failed to create analytics country index: %w", err)
		}
	}

	// Add composite index for transcripts
	if !db.Migrator().HasIndex(&models.Transcript{}, "idx_episode_lang") {
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_episode_lang ON transcripts(episode_id, language)").Error; err != nil {
			return fmt.Errorf("failed to create transcript composite index: %w", err)
		}
	}

	logger.Logger.Info().Msg("Migration 005 completed successfully")
	return nil
}

func migration005Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 005: add_analytics_and_transcripts")

	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_analytics_episode_id")
	db.Exec("DROP INDEX IF EXISTS idx_analytics_last_played")
	db.Exec("DROP INDEX IF EXISTS idx_analytics_country")
	db.Exec("DROP INDEX IF EXISTS idx_episode_lang")

	// Drop tables
	if err := db.Migrator().DropTable(&models.Transcript{}); err != nil {
		return err
	}

	if err := db.Migrator().DropTable(&models.Analytics{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 005 rollback completed")
	return nil
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *gorm.DB) ([]map[string]interface{}, error) {
	currentVersion := getCurrentVersion(db)

	var executedMigrations []SchemaMigration
	if err := db.Order("version ASC").Find(&executedMigrations).Error; err != nil {
		return nil, err
	}

	executedMap := make(map[int]SchemaMigration)
	for _, m := range executedMigrations {
		executedMap[m.Version] = m
	}

	status := make([]map[string]interface{}, 0, len(migrations))
	for _, migration := range migrations {
		entry := map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
			"status":  "pending",
		}

		if executed, ok := executedMap[migration.Version]; ok {
			entry["status"] = "executed"
			entry["executed_at"] = executed.ExecutedAt
		}

		if migration.Version == currentVersion {
			entry["is_current"] = true
		}

		status = append(status, entry)
	}

	return status, nil
}

// migration006Up adds webhook and batch job tables
func migration006Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 006: add_webhooks_and_batch_jobs")

	// Create WebhookConfig table
	if err := db.AutoMigrate(&models.WebhookConfig{}); err != nil {
		return fmt.Errorf("failed to migrate WebhookConfig table: %w", err)
	}

	// Create WebhookDelivery table
	if err := db.AutoMigrate(&models.WebhookDelivery{}); err != nil {
		return fmt.Errorf("failed to migrate WebhookDelivery table: %w", err)
	}

	// Create BatchJobStatus table
	if err := db.AutoMigrate(&models.BatchJobStatus{}); err != nil {
		return fmt.Errorf("failed to migrate BatchJobStatus table: %w", err)
	}

	// Add indexes for webhook_deliveries
	if !db.Migrator().HasIndex(&models.WebhookDelivery{}, "idx_webhook_config_id") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_webhook_config_id ON webhook_deliveries(webhook_config_id)").Error; err != nil {
			return fmt.Errorf("failed to create webhook_config_id index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.WebhookDelivery{}, "idx_webhook_status") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_webhook_status ON webhook_deliveries(status)").Error; err != nil {
			return fmt.Errorf("failed to create webhook status index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.WebhookDelivery{}, "idx_webhook_next_retry") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_webhook_next_retry ON webhook_deliveries(next_retry_at)").Error; err != nil {
			return fmt.Errorf("failed to create webhook next_retry_at index: %w", err)
		}
	}

	// Add indexes for batch_job_statuses
	if !db.Migrator().HasIndex(&models.BatchJobStatus{}, "idx_batch_job_type") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_batch_job_type ON batch_job_statuses(job_type)").Error; err != nil {
			return fmt.Errorf("failed to create batch job_type index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.BatchJobStatus{}, "idx_batch_status") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_batch_status ON batch_job_statuses(status)").Error; err != nil {
			return fmt.Errorf("failed to create batch status index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.BatchJobStatus{}, "idx_batch_created_at") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_batch_created_at ON batch_job_statuses(created_at)").Error; err != nil {
			return fmt.Errorf("failed to create batch created_at index: %w", err)
		}
	}

	logger.Logger.Info().Msg("Migration 006 completed successfully")
	return nil
}

func migration006Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 006: add_webhooks_and_batch_jobs")

	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_webhook_config_id")
	db.Exec("DROP INDEX IF EXISTS idx_webhook_status")
	db.Exec("DROP INDEX IF EXISTS idx_webhook_next_retry")
	db.Exec("DROP INDEX IF EXISTS idx_batch_job_type")
	db.Exec("DROP INDEX IF EXISTS idx_batch_status")
	db.Exec("DROP INDEX IF EXISTS idx_batch_created_at")

	// Drop tables
	if err := db.Migrator().DropTable(&models.BatchJobStatus{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.WebhookDelivery{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.WebhookConfig{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 006 rollback completed")
	return nil
}

// migration007Up adds download progress table
func migration007Up(db *gorm.DB) error {
	logger.Logger.Info().Msg("Running migration 007: add_download_progress")

	// Create DownloadProgress table
	if err := db.AutoMigrate(&models.DownloadProgress{}); err != nil {
		return fmt.Errorf("failed to migrate DownloadProgress table: %w", err)
	}

	// Add indexes for download_progress
	if !db.Migrator().HasIndex(&models.DownloadProgress{}, "idx_download_video_id") {
		if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_download_video_id ON download_progresses(video_id)").Error; err != nil {
			return fmt.Errorf("failed to create download video_id index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.DownloadProgress{}, "idx_download_status") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_download_status ON download_progresses(status)").Error; err != nil {
			return fmt.Errorf("failed to create download status index: %w", err)
		}
	}

	if !db.Migrator().HasIndex(&models.DownloadProgress{}, "idx_download_started_at") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_download_started_at ON download_progresses(started_at)").Error; err != nil {
			return fmt.Errorf("failed to create download started_at index: %w", err)
		}
	}

	logger.Logger.Info().Msg("Migration 007 completed successfully")
	return nil
}

func migration007Down(db *gorm.DB) error {
	logger.Logger.Info().Msg("Rolling back migration 007: add_download_progress")

	// Drop indexes
	db.Exec("DROP INDEX IF EXISTS idx_download_video_id")
	db.Exec("DROP INDEX IF EXISTS idx_download_status")
	db.Exec("DROP INDEX IF EXISTS idx_download_started_at")

	// Drop table
	if err := db.Migrator().DropTable(&models.DownloadProgress{}); err != nil {
		return err
	}

	logger.Logger.Info().Msg("Migration 007 rollback completed")
	return nil
}
