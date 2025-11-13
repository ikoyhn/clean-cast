package downloader

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/metrics"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/webhook"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lrstanley/go-ytdlp"
)

var (
	youtubeVideoMutexes = &sync.Map{}
	onDownloadStart     func()
	onDownloadComplete  func()
	// Semaphore to limit concurrent downloads to 5
	downloadSemaphore = make(chan struct{}, 5)
)

const (
	youtubeVideoUrl = "https://www.youtube.com/watch?v="
	maxRetryAttempts = 3
	tempDownloadDir  = "temp_downloads"
)

// AudioFormat represents the audio format for downloads
type AudioFormat struct {
	Format      string // m4a, mp3, opus, etc.
	Quality     string // 128k, 192k, 320k
	Extension   string // File extension
	MimeType    string // MIME type for HTTP responses
	FormatSort  string // yt-dlp format sort string
}

// Predefined audio formats
var (
	FormatM4A = AudioFormat{
		Format:     "m4a",
		Quality:    "192k",
		Extension:  ".m4a",
		MimeType:   "audio/mp4",
		FormatSort: "ext::m4a[format_note*=original]",
	}
	FormatMP3 = AudioFormat{
		Format:     "mp3",
		Quality:    "192k",
		Extension:  ".mp3",
		MimeType:   "audio/mpeg",
		FormatSort: "ext",
	}
	FormatOpus = AudioFormat{
		Format:     "opus",
		Quality:    "128k",
		Extension:  ".opus",
		MimeType:   "audio/opus",
		FormatSort: "ext",
	}
)

// GetAudioFormat returns the audio format based on format string
func GetAudioFormat(format string) AudioFormat {
	switch strings.ToLower(format) {
	case "mp3":
		return FormatMP3
	case "opus":
		return FormatOpus
	case "m4a", "":
		return FormatM4A
	default:
		return FormatM4A
	}
}

// SetDownloadCallbacks sets the callbacks for download start and completion
func SetDownloadCallbacks(onStart, onComplete func()) {
	onDownloadStart = onStart
	onDownloadComplete = onComplete
}

func GetYoutubeVideo(youtubeVideoId string) (string, <-chan struct{}) {
	return GetYoutubeVideoWithFormat(youtubeVideoId, FormatM4A)
}

// GetYoutubeVideoWithFormat downloads a YouTube video with specified format and quality
func GetYoutubeVideoWithFormat(youtubeVideoId string, audioFormat AudioFormat) (string, <-chan struct{}) {
	// Strip any existing extension from the video ID
	for _, format := range []AudioFormat{FormatM4A, FormatMP3, FormatOpus} {
		youtubeVideoId = strings.TrimSuffix(youtubeVideoId, format.Extension)
	}

	mutex, ok := youtubeVideoMutexes.Load(youtubeVideoId)
	if !ok {
		mutex = &sync.Mutex{}
		youtubeVideoMutexes.Store(youtubeVideoId, mutex)
	}

	mutex.(*sync.Mutex).Lock()

	// Check if the file is already being processed
	filePath := filepath.Join(config.Config.AudioDir, youtubeVideoId+audioFormat.Extension)
	if _, err := os.Stat(filePath); err == nil {
		mutex.(*sync.Mutex).Unlock()
		done := make(chan struct{})
		close(done)
		return youtubeVideoId, done
	}

	// Get audio quality for logging
	quality := audioFormat.Quality
	if config.Config.AudioQuality != "" {
		quality = config.Config.AudioQuality
	}

	done := make(chan struct{})
	go func() {
		startTime := time.Now()

		// Acquire semaphore slot (blocks if 5 downloads are already running)
		downloadSemaphore <- struct{}{}
		logger.Logger.Debug().
			Str("video_id", youtubeVideoId).
			Msg("Acquired download slot")

		// Track download start
		metrics.IncActiveDownloads()
		if onDownloadStart != nil {
			onDownloadStart()
		}

		defer func() {
			// Track download completion
			duration := time.Since(startTime)
			metrics.RecordDownload(duration)
			metrics.DecActiveDownloads()

			if onDownloadComplete != nil {
				onDownloadComplete()
			}
			mutex.(*sync.Mutex).Unlock()
			// Clean up mutex from map to prevent memory leak
			youtubeVideoMutexes.Delete(youtubeVideoId)
			// Release semaphore slot
			<-downloadSemaphore
			logger.Logger.Debug().
				Str("video_id", youtubeVideoId).
				Dur("duration", duration).
				Msg("Released download slot")
			close(done)
		}()

		// Use the new download with retry logic
		ctx := context.Background()
		dlErr := downloadWithRetry(ctx, youtubeVideoId, audioFormat)

		if dlErr != nil {
			metrics.RecordError("download_failed")
			logger.Logger.Error().
				Err(dlErr).
				Str("video_id", youtubeVideoId).
				Str("format", audioFormat.Format).
				Msg("Error downloading YouTube video")

			// Send webhook notification for download error
			webhook.SendWebhook(webhook.EventError, map[string]interface{}{
				"error":    dlErr.Error(),
				"video_id": youtubeVideoId,
				"format":   audioFormat.Format,
				"context":  "download_failed",
			})
		} else {
			logger.Logger.Info().
				Str("video_id", youtubeVideoId).
				Str("format", audioFormat.Format).
				Str("quality", quality).
				Dur("duration", time.Since(startTime)).
				Msg("Download completed successfully")

			// Send webhook notification for download complete
			webhook.SendWebhook(webhook.EventDownloadComplete, map[string]interface{}{
				"video_id":  youtubeVideoId,
				"format":    audioFormat.Format,
				"quality":   quality,
				"duration":  time.Since(startTime).String(),
				"file_path": youtubeVideoId + audioFormat.Extension,
			})
		}
	}()

	return youtubeVideoId, done
}

// ensureTempDir ensures the temporary download directory exists
func ensureTempDir() (string, error) {
	tempDir := filepath.Join(config.Config.AudioDir, tempDownloadDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("temp_dir", tempDir).
			Msg("Failed to create temporary download directory")
		return "", err
	}
	return tempDir, nil
}

// getTempFilePath returns the temporary file path for a download
func getTempFilePath(videoId string, audioFormat AudioFormat) (string, error) {
	tempDir, err := ensureTempDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tempDir, videoId+audioFormat.Extension+".part"), nil
}

// getFinalFilePath returns the final file path for a download
func getFinalFilePath(videoId string, audioFormat AudioFormat) string {
	return filepath.Join(config.Config.AudioDir, videoId+audioFormat.Extension)
}

// moveToFinalLocation moves a file from temporary location to final location
func moveToFinalLocation(tempPath, finalPath string) error {
	if err := os.Rename(tempPath, finalPath); err != nil {
		logger.Logger.Error().
			Err(err).
			Str("temp_path", tempPath).
			Str("final_path", finalPath).
			Msg("Failed to move file to final location")
		return err
	}

	logger.Logger.Info().
		Str("temp_path", tempPath).
		Str("final_path", finalPath).
		Msg("Moved download to final location")

	return nil
}

// cleanupTempFile removes a temporary file
func cleanupTempFile(tempPath string) {
	if tempPath == "" {
		return
	}

	if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
		logger.Logger.Warn().
			Err(err).
			Str("temp_path", tempPath).
			Msg("Failed to cleanup temporary file")
	}
}

// downloadWithRetry performs a download with retry logic
func downloadWithRetry(ctx context.Context, videoId string, audioFormat AudioFormat) error {
	// Get or create download progress
	progress, err := database.GetOrCreateDownloadProgress(videoId, audioFormat.Format)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Failed to get or create download progress")
		return err
	}

	// Check if already completed
	if progress.IsComplete() {
		logger.Logger.Info().
			Str("video_id", videoId).
			Msg("Download already completed")
		return nil
	}

	// Check if can retry
	if !progress.CanRetry(maxRetryAttempts) {
		err := fmt.Errorf("max retry attempts (%d) exceeded", maxRetryAttempts)
		logger.Logger.Error().
			Str("video_id", videoId).
			Int("retry_count", progress.RetryCount).
			Msg("Cannot retry download")
		return err
	}

	// Get paths
	tempPath, err := getTempFilePath(videoId, audioFormat)
	if err != nil {
		return err
	}
	finalPath := getFinalFilePath(videoId, audioFormat)

	// Configure retry settings
	retryConfig := common.DownloadRetryConfig()
	retryConfig.OnRetry = func(attempt int, err error) {
		logger.Logger.Warn().
			Str("video_id", videoId).
			Int("attempt", attempt).
			Err(err).
			Msg("Download attempt failed, retrying...")

		// Update progress in database
		database.MarkDownloadFailed(videoId, err)
	}

	// Perform download with retry
	err = common.RetryWithBackoff(ctx, retryConfig, func() error {
		return performDownload(ctx, videoId, audioFormat, tempPath, finalPath, progress)
	})

	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Download failed after all retries")
		database.MarkDownloadFailed(videoId, err)
		cleanupTempFile(tempPath)
		return err
	}

	// Mark as completed
	err = database.MarkDownloadCompleted(videoId, finalPath)
	if err != nil {
		logger.Logger.Warn().
			Err(err).
			Str("video_id", videoId).
			Msg("Failed to mark download as completed in database")
	}

	return nil
}

// performDownload performs the actual download
func performDownload(ctx context.Context, videoId string, audioFormat AudioFormat, tempPath, finalPath string, progress *models.DownloadProgress) error {
	// Mark as in progress
	err := database.MarkDownloadInProgress(videoId, tempPath)
	if err != nil {
		logger.Logger.Warn().
			Err(err).
			Str("video_id", videoId).
			Msg("Failed to mark download as in progress")
	}

	// Ensure ytdlp is installed
	ytdlp.Install(ctx, nil)

	// Get SponsorBlock categories
	categories := config.Config.SponsorBlockCategories
	if categories == "" {
		categories = "sponsor"
	}
	categories = strings.TrimSpace(categories)

	// Configure audio quality
	quality := audioFormat.Quality
	if config.Config.AudioQuality != "" {
		quality = config.Config.AudioQuality
	}

	// Create temporary directory for download
	tempDir := filepath.Dir(tempPath)

	// Configure ytdlp
	dl := ytdlp.New().
		NoProgress().
		FormatSort(audioFormat.FormatSort).
		SponsorblockRemove(categories).
		ExtractAudio().
		AudioFormat(audioFormat.Format).
		AudioQuality(quality).
		NoPlaylist().
		FFmpegLocation("/usr/bin/ffmpeg").
		Continue().
		Paths(tempDir).
		Output(videoId + ".%(ext)s")

	if config.Config.CookiesFile != "" {
		dl.Cookies(config.Config.CookiesFile)
	}

	// Run download
	r, dlErr := dl.Run(ctx, youtubeVideoUrl+videoId)

	if r.ExitCode != 0 {
		// Check if file exists despite non-zero exit code
		downloadedFile := filepath.Join(tempDir, videoId+audioFormat.Extension)
		if _, err := os.Stat(downloadedFile); err == nil {
			logger.Logger.Warn().
				Str("video_id", videoId).
				Str("format", audioFormat.Format).
				Int("exit_code", r.ExitCode).
				Msg("Download exited with non-zero code, but file exists")

			// Move to final location
			return moveToFinalLocation(downloadedFile, finalPath)
		}

		if dlErr != nil {
			metrics.RecordError("download_failed")
			return fmt.Errorf("download failed with exit code %d: %w", r.ExitCode, dlErr)
		}

		return fmt.Errorf("download failed with exit code %d", r.ExitCode)
	}

	// Move completed download to final location
	downloadedFile := filepath.Join(tempDir, videoId+audioFormat.Extension)
	if err := moveToFinalLocation(downloadedFile, finalPath); err != nil {
		return err
	}

	logger.Logger.Info().
		Str("video_id", videoId).
		Str("format", audioFormat.Format).
		Str("quality", quality).
		Msg("Download completed successfully")

	return nil
}
