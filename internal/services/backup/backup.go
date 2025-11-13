package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type BackupMetadata struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`
	Description string    `json:"description"`
}

type BackupData struct {
	Metadata                BackupMetadata                     `json:"metadata"`
	Podcasts                []models.Podcast                   `json:"podcasts"`
	Episodes                []models.PodcastEpisode            `json:"episodes"`
	EpisodePlaybackHistory  []models.EpisodePlaybackHistory    `json:"episode_playback_history"`
	AudioFiles              []string                           `json:"audio_files"`
}

// GetBackupDir returns the backup directory path
func GetBackupDir() string {
	backupDir := filepath.Join(config.Config.ConfigDir, "backups")
	os.MkdirAll(backupDir, os.ModePerm)
	return backupDir
}

// ExportDatabase exports all database tables to JSON
func ExportDatabase() (*BackupData, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	backup := &BackupData{
		Metadata: BackupMetadata{
			ID:          fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405")),
			CreatedAt:   time.Now(),
			Description: "Full database backup",
		},
	}

	// Export podcasts
	if err := db.Find(&backup.Podcasts).Error; err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to export podcasts")
		return nil, err
	}

	// Export episodes
	if err := db.Find(&backup.Episodes).Error; err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to export episodes")
		return nil, err
	}

	// Export playback history
	if err := db.Find(&backup.EpisodePlaybackHistory).Error; err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to export playback history")
		return nil, err
	}

	// Get list of audio files
	audioFiles, err := GetAudioFilesList()
	if err != nil {
		logger.Logger.Warn().Err(err).Msg("Failed to get audio files list")
	}
	backup.AudioFiles = audioFiles
	backup.Metadata.FileCount = len(audioFiles)

	logger.Logger.Info().
		Int("podcasts", len(backup.Podcasts)).
		Int("episodes", len(backup.Episodes)).
		Int("audio_files", len(backup.AudioFiles)).
		Msg("Database exported successfully")

	return backup, nil
}

// GetAudioFilesList returns a list of all audio files in the audio directory
func GetAudioFilesList() ([]string, error) {
	var audioFiles []string

	if _, err := os.Stat(config.Config.AudioDir); os.IsNotExist(err) {
		return audioFiles, nil
	}

	files, err := os.ReadDir(config.Config.AudioDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			audioFiles = append(audioFiles, file.Name())
		}
	}

	return audioFiles, nil
}

// CreateBackup creates a full backup including database and optionally audio files
func CreateBackup(includeAudio bool, description string) (string, error) {
	// Export database
	backupData, err := ExportDatabase()
	if err != nil {
		return "", err
	}

	if description != "" {
		backupData.Metadata.Description = description
	}

	backupDir := GetBackupDir()
	backupID := backupData.Metadata.ID
	backupPath := filepath.Join(backupDir, backupID+".zip")

	// Create zip file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add database JSON to zip
	dbJSON, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return "", err
	}

	dbFile, err := zipWriter.Create("database.json")
	if err != nil {
		return "", err
	}
	if _, err := dbFile.Write(dbJSON); err != nil {
		return "", err
	}

	// Add audio files if requested
	if includeAudio {
		logger.Logger.Info().Msg("Including audio files in backup")
		for _, audioFile := range backupData.AudioFiles {
			audioPath := filepath.Join(config.Config.AudioDir, audioFile)
			if err := addFileToZip(zipWriter, audioPath, "audio/"+audioFile); err != nil {
				logger.Logger.Warn().
					Err(err).
					Str("file", audioFile).
					Msg("Failed to add audio file to backup")
				continue
			}
		}
	}

	// Get file size
	fileInfo, err := os.Stat(backupPath)
	if err == nil {
		backupData.Metadata.Size = fileInfo.Size()
	}

	// Save metadata
	metadataPath := filepath.Join(backupDir, backupID+".json")
	metadataJSON, err := json.MarshalIndent(backupData.Metadata, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		logger.Logger.Warn().Err(err).Msg("Failed to save backup metadata")
	}

	logger.Logger.Info().
		Str("backup_id", backupID).
		Str("path", backupPath).
		Int64("size", backupData.Metadata.Size).
		Msg("Backup created successfully")

	return backupID, nil
}

// addFileToZip adds a file to the zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// ListBackups returns a list of all available backups
func ListBackups() ([]BackupMetadata, error) {
	backupDir := GetBackupDir()
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []BackupMetadata
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			metadataPath := filepath.Join(backupDir, file.Name())
			data, err := os.ReadFile(metadataPath)
			if err != nil {
				logger.Logger.Warn().
					Err(err).
					Str("file", file.Name()).
					Msg("Failed to read backup metadata")
				continue
			}

			var metadata BackupMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				logger.Logger.Warn().
					Err(err).
					Str("file", file.Name()).
					Msg("Failed to parse backup metadata")
				continue
			}

			backups = append(backups, metadata)
		}
	}

	return backups, nil
}

// RestoreFromBackup restores database and optionally audio files from a backup
func RestoreFromBackup(backupID string, includeAudio bool) error {
	backupDir := GetBackupDir()
	backupPath := filepath.Join(backupDir, backupID+".zip")

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Open zip file
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Find and read database.json
	var backupData BackupData
	for _, file := range zipReader.File {
		if file.Name == "database.json" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(data, &backupData); err != nil {
				return err
			}
			break
		}
	}

	// Restore database
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Begin transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Clear existing data
	if err := tx.Where("1 = 1").Delete(&models.Podcast{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("1 = 1").Delete(&models.PodcastEpisode{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("1 = 1").Delete(&models.EpisodePlaybackHistory{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Restore podcasts
	if len(backupData.Podcasts) > 0 {
		if err := tx.Create(&backupData.Podcasts).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Restore episodes
	if len(backupData.Episodes) > 0 {
		if err := tx.Create(&backupData.Episodes).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Restore playback history
	if len(backupData.EpisodePlaybackHistory) > 0 {
		if err := tx.Create(&backupData.EpisodePlaybackHistory).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Restore audio files if requested
	if includeAudio {
		logger.Logger.Info().Msg("Restoring audio files")
		for _, file := range zipReader.File {
			if filepath.Dir(file.Name) == "audio" {
				if err := extractFileFromZip(file, config.Config.AudioDir); err != nil {
					logger.Logger.Warn().
						Err(err).
						Str("file", file.Name).
						Msg("Failed to restore audio file")
					continue
				}
			}
		}
	}

	logger.Logger.Info().
		Str("backup_id", backupID).
		Int("podcasts", len(backupData.Podcasts)).
		Int("episodes", len(backupData.Episodes)).
		Msg("Backup restored successfully")

	return nil
}

// extractFileFromZip extracts a single file from zip to destination
func extractFileFromZip(file *zip.File, destDir string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	fileName := filepath.Base(file.Name)
	destPath := filepath.Join(destDir, fileName)

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, rc)
	return err
}

// DeleteBackup deletes a backup by ID
func DeleteBackup(backupID string) error {
	backupDir := GetBackupDir()

	// Delete zip file
	zipPath := filepath.Join(backupDir, backupID+".zip")
	if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Delete metadata file
	metadataPath := filepath.Join(backupDir, backupID+".json")
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	logger.Logger.Info().Str("backup_id", backupID).Msg("Backup deleted successfully")
	return nil
}

// UploadToS3 uploads a backup to S3 (optional cloud storage)
func UploadToS3(backupID, bucket, region, accessKey, secretKey string) error {
	backupDir := GetBackupDir()
	backupPath := filepath.Join(backupDir, backupID+".zip")

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create S3 session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return err
	}

	// Upload to S3
	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(backupID + ".zip"),
		Body:   file,
	})

	if err != nil {
		return err
	}

	logger.Logger.Info().
		Str("backup_id", backupID).
		Str("bucket", bucket).
		Msg("Backup uploaded to S3 successfully")

	return nil
}

// DownloadFromS3 downloads a backup from S3
func DownloadFromS3(backupID, bucket, region, accessKey, secretKey string) error {
	backupDir := GetBackupDir()
	backupPath := filepath.Join(backupDir, backupID+".zip")

	// Create S3 session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return err
	}

	// Download from S3
	svc := s3.New(sess)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(backupID + ".zip"),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	// Create local file
	file, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return err
	}

	logger.Logger.Info().
		Str("backup_id", backupID).
		Str("bucket", bucket).
		Msg("Backup downloaded from S3 successfully")

	return nil
}

// ScheduledBackup performs a scheduled backup (can be called by cron)
func ScheduledBackup(includeAudio bool) error {
	backupID, err := CreateBackup(includeAudio, "Scheduled backup")
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Scheduled backup failed")
		return err
	}

	logger.Logger.Info().
		Str("backup_id", backupID).
		Msg("Scheduled backup completed successfully")

	// Upload to S3 if configured
	s3Bucket := os.Getenv("BACKUP_S3_BUCKET")
	s3Region := os.Getenv("BACKUP_S3_REGION")
	s3AccessKey := os.Getenv("BACKUP_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("BACKUP_S3_SECRET_KEY")

	if s3Bucket != "" && s3Region != "" && s3AccessKey != "" && s3SecretKey != "" {
		if err := UploadToS3(backupID, s3Bucket, s3Region, s3AccessKey, s3SecretKey); err != nil {
			logger.Logger.Error().
				Err(err).
				Str("backup_id", backupID).
				Msg("Failed to upload backup to S3")
		}
	}

	return nil
}
