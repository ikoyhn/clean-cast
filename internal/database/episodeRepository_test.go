package database

import (
	"os"
	"path"
	"testing"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

func setupTestDB(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	config.Config.ConfigDir = tmpDir
	config.Config.DbFile = path.Join(tmpDir, "test.db")
	config.Config.AudioDir = path.Join(tmpDir, "audio")

	if err := os.MkdirAll(config.Config.AudioDir, 0755); err != nil {
		t.Fatalf("failed to create audio dir: %v", err)
	}

	SetupDatabase()

	return tmpDir
}

func TestDeletePodcastCronJob_RemovesExistingFileAndRecord(t *testing.T) {
	tmp := setupTestDB(t)

	videoId := "video1"
	filePath := path.Join(config.Config.AudioDir, videoId+".m4a")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hist := &models.EpisodePlaybackHistory{
		YoutubeVideoId:   videoId,
		LastAccessDate:   time.Now().Add(-8 * 24 * time.Hour).Unix(),
		TotalTimeSkipped: 0,
	}
	if err := db.Create(hist).Error; err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	DeletePodcastCronJob()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat error: %v", err)
	}

	var out models.EpisodePlaybackHistory
	if err := db.Where("youtube_video_id = ?", videoId).First(&out).Error; err == nil {
		t.Fatalf("expected DB record to be deleted")
	}

	_ = tmp
}

func TestDeletePodcastCronJob_DeletesRecordWhenFileMissing(t *testing.T) {
	setupTestDB(t)

	videoId := "video-missing"
	hist := &models.EpisodePlaybackHistory{
		YoutubeVideoId:   videoId,
		LastAccessDate:   time.Now().Add(-8 * 24 * time.Hour).Unix(),
		TotalTimeSkipped: 0,
	}
	if err := db.Create(hist).Error; err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	DeletePodcastCronJob()

	var out models.EpisodePlaybackHistory
	if err := db.Where("youtube_video_id = ?", videoId).First(&out).Error; err == nil {
		t.Fatalf("expected DB record to be deleted for missing file")
	}
}
