package tests

import (
	"os"
	"path"
	"testing"

	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
)

func SetupIntegration(t *testing.T) {
	t.Helper()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set; skipping integration test")
	}

	tmpDir := t.TempDir()
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	config.AppConfig.Setup.GoogleApiKey = apiKey
	config.AppConfig.Setup.ConfigDir = tmpDir
	config.AppConfig.Setup.DbFile = path.Join(tmpDir, "test.db")
	config.AppConfig.Setup.AudioDir = path.Join(tmpDir, "audio")
	config.AppConfig.Setup.PodcastRefreshInterval = "0s"
	config.AppConfig.Ytdlp.EpisodeDurationMinimum = "0s"

	if err := os.MkdirAll(config.AppConfig.Setup.AudioDir, 0755); err != nil {
		t.Fatalf("failed to create audio dir: %v", err)
	}

	database.SetupDatabase()
	youtube.SetupYoutubeService()
}
