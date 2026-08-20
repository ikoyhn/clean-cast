package database

import (
	"os"
	"path"
	"testing"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

func TestTrackEpisodeFiles_KeepsHistoryForLongExtensions(t *testing.T) {
	setupTestDB(t)

	videoId := "video-opus-1"
	filePath := path.Join(config.AppConfig.Setup.AudioDir, videoId+".opus")
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

	TrackEpisodeFiles()

	var out models.EpisodePlaybackHistory
	if err := db.Where("youtube_video_id = ?", videoId).First(&out).Error; err != nil {
		t.Fatalf("expected history for .opus file to survive TrackEpisodeFiles: %v", err)
	}
	if out.LastAccessDate != hist.LastAccessDate {
		t.Fatalf("expected LastAccessDate to be unchanged")
	}
}

func TestTrackEpisodeFiles_AdoptsUntrackedFile(t *testing.T) {
	setupTestDB(t)

	videoId := "video-opus-2"
	filePath := path.Join(config.AppConfig.Setup.AudioDir, videoId+".opus")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	TrackEpisodeFiles()

	var out models.EpisodePlaybackHistory
	if err := db.Where("youtube_video_id = ?", videoId).First(&out).Error; err != nil {
		t.Fatalf("expected history row to be created for untracked .opus file: %v", err)
	}
}

func TestUpdateEpisodePlaybackHistory_RefreshesLastAccessDate(t *testing.T) {
	setupTestDB(t)

	videoId := "video-refresh"
	staleDate := time.Now().Add(-30 * 24 * time.Hour).Unix()
	hist := &models.EpisodePlaybackHistory{
		YoutubeVideoId:   videoId,
		LastAccessDate:   staleDate,
		TotalTimeSkipped: 5,
	}
	if err := db.Create(hist).Error; err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	UpdateEpisodePlaybackHistory(videoId, 12)

	var out models.EpisodePlaybackHistory
	if err := db.Where("youtube_video_id = ?", videoId).First(&out).Error; err != nil {
		t.Fatalf("failed to read history: %v", err)
	}
	if out.LastAccessDate == staleDate {
		t.Fatalf("expected LastAccessDate to be refreshed on access")
	}
	if out.TotalTimeSkipped != 12 {
		t.Fatalf("expected TotalTimeSkipped to be updated, got %v", out.TotalTimeSkipped)
	}
}
