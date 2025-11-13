package api

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/services/backup"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/middleware"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// RegisterBackupRoutes registers all backup-related routes
func RegisterBackupRoutes(e *echo.Echo) {
	// Backup API endpoints - require auth and moderate rate limiting
	api := e.Group("/api/backup")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.RateLimitMiddleware(30))

	api.POST("/create", createBackupHandler)
	api.GET("/list", listBackupsHandler)
	api.POST("/restore", restoreBackupHandler)
	api.GET("/download/:id", downloadBackupHandler)
	api.DELETE("/:id", deleteBackupHandler)
	api.POST("/upload-s3", uploadToS3Handler)
	api.POST("/download-s3", downloadFromS3Handler)
}

// CreateBackupRequest represents the request body for creating a backup
type CreateBackupRequest struct {
	IncludeAudio bool   `json:"include_audio"`
	Description  string `json:"description"`
}

// CreateBackupResponse represents the response for backup creation
type CreateBackupResponse struct {
	BackupID string `json:"backup_id"`
	Message  string `json:"message"`
}

// RestoreBackupRequest represents the request body for restoring a backup
type RestoreBackupRequest struct {
	BackupID     string `json:"backup_id"`
	IncludeAudio bool   `json:"include_audio"`
}

// S3Request represents the request body for S3 operations
type S3Request struct {
	BackupID  string `json:"backup_id"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// createBackupHandler handles POST /api/backup/create
func createBackupHandler(c echo.Context) error {
	var req CreateBackupRequest
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	logger.Logger.Info().
		Bool("include_audio", req.IncludeAudio).
		Str("description", req.Description).
		Msg("Creating backup")

	backupID, err := backup.CreateBackup(req.IncludeAudio, req.Description)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create backup")
		return appErrors.NewInternalServerError("Failed to create backup").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, CreateBackupResponse{
		BackupID: backupID,
		Message:  "Backup created successfully",
	})
}

// listBackupsHandler handles GET /api/backup/list
func listBackupsHandler(c echo.Context) error {
	backups, err := backup.ListBackups()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to list backups")
		return appErrors.NewInternalServerError("Failed to list backups").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"backups": backups,
		"count":   len(backups),
	})
}

// restoreBackupHandler handles POST /api/backup/restore
func restoreBackupHandler(c echo.Context) error {
	var req RestoreBackupRequest
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	if req.BackupID == "" {
		return appErrors.NewBadRequestError("backup_id is required")
	}

	logger.Logger.Info().
		Str("backup_id", req.BackupID).
		Bool("include_audio", req.IncludeAudio).
		Msg("Restoring backup")

	if err := backup.RestoreFromBackup(req.BackupID, req.IncludeAudio); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("backup_id", req.BackupID).
			Msg("Failed to restore backup")
		return appErrors.NewInternalServerError("Failed to restore backup").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup restored successfully",
	})
}

// downloadBackupHandler handles GET /api/backup/download/:id
func downloadBackupHandler(c echo.Context) error {
	backupID := c.Param("id")
	if backupID == "" {
		return appErrors.NewBadRequestError("backup_id is required")
	}

	backupDir := backup.GetBackupDir()
	backupPath := filepath.Join(backupDir, backupID+".zip")

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return appErrors.NewFileNotFoundError(backupID + ".zip")
	}

	// Set headers for download
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", backupID))
	c.Response().Header().Set("Content-Type", "application/zip")

	return c.File(backupPath)
}

// deleteBackupHandler handles DELETE /api/backup/:id
func deleteBackupHandler(c echo.Context) error {
	backupID := c.Param("id")
	if backupID == "" {
		return appErrors.NewBadRequestError("backup_id is required")
	}

	logger.Logger.Info().Str("backup_id", backupID).Msg("Deleting backup")

	if err := backup.DeleteBackup(backupID); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("backup_id", backupID).
			Msg("Failed to delete backup")
		return appErrors.NewInternalServerError("Failed to delete backup").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup deleted successfully",
	})
}

// uploadToS3Handler handles POST /api/backup/upload-s3
func uploadToS3Handler(c echo.Context) error {
	var req S3Request
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	// Validate required fields
	if req.BackupID == "" || req.Bucket == "" || req.Region == "" {
		return appErrors.NewBadRequestError("backup_id, bucket, and region are required")
	}

	// Use environment variables if credentials not provided
	accessKey := req.AccessKey
	secretKey := req.SecretKey
	if accessKey == "" {
		accessKey = os.Getenv("BACKUP_S3_ACCESS_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("BACKUP_S3_SECRET_KEY")
	}

	if accessKey == "" || secretKey == "" {
		return appErrors.NewBadRequestError("S3 credentials not provided")
	}

	logger.Logger.Info().
		Str("backup_id", req.BackupID).
		Str("bucket", req.Bucket).
		Msg("Uploading backup to S3")

	if err := backup.UploadToS3(req.BackupID, req.Bucket, req.Region, accessKey, secretKey); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("backup_id", req.BackupID).
			Msg("Failed to upload backup to S3")
		return appErrors.NewInternalServerError("Failed to upload backup to S3").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup uploaded to S3 successfully",
	})
}

// downloadFromS3Handler handles POST /api/backup/download-s3
func downloadFromS3Handler(c echo.Context) error {
	var req S3Request
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	// Validate required fields
	if req.BackupID == "" || req.Bucket == "" || req.Region == "" {
		return appErrors.NewBadRequestError("backup_id, bucket, and region are required")
	}

	// Use environment variables if credentials not provided
	accessKey := req.AccessKey
	secretKey := req.SecretKey
	if accessKey == "" {
		accessKey = os.Getenv("BACKUP_S3_ACCESS_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("BACKUP_S3_SECRET_KEY")
	}

	if accessKey == "" || secretKey == "" {
		return appErrors.NewBadRequestError("S3 credentials not provided")
	}

	logger.Logger.Info().
		Str("backup_id", req.BackupID).
		Str("bucket", req.Bucket).
		Msg("Downloading backup from S3")

	if err := backup.DownloadFromS3(req.BackupID, req.Bucket, req.Region, accessKey, secretKey); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("backup_id", req.BackupID).
			Msg("Failed to download backup from S3")
		return appErrors.NewInternalServerError("Failed to download backup from S3").
			WithDetail("error", err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup downloaded from S3 successfully",
	})
}
