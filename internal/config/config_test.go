package config

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func writeConfigDir(t *testing.T, properties string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "properties.yml"), []byte(properties), 0644); err != nil {
		t.Fatalf("failed to write properties.yml: %v", err)
	}
	t.Setenv("CONFIG_DIR", dir)
	return dir
}

func TestLoad_ReadsPodcastRefreshIntervalFromProperties(t *testing.T) {
	writeConfigDir(t, `
setup:
    google-api-key: testkey
    podcast-refresh-interval: 30s
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Setup.PodcastRefreshInterval != 30*time.Second {
		t.Errorf("refresh interval = %v, want 30s", cfg.Setup.PodcastRefreshInterval)
	}
}

func TestLoad_DefaultsWhenKeyPresentButEmpty(t *testing.T) {
	// Matches the shipped template, where keys are listed with no value.
	writeConfigDir(t, `
setup:
    google-api-key: testkey
    cron:
    podcast-refresh-interval:
ytdlp:
    episode-duration-minimum:
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Setup.PodcastRefreshInterval != time.Hour {
		t.Errorf("refresh interval = %v, want 1h", cfg.Setup.PodcastRefreshInterval)
	}
	if cfg.Ytdlp.EpisodeDurationMinimum != 3*time.Minute {
		t.Errorf("duration minimum = %v, want 3m", cfg.Ytdlp.EpisodeDurationMinimum)
	}
}

func TestLoad_RejectsInvalidDurationInsteadOfPanicking(t *testing.T) {
	// "1hr" is not valid Go duration syntax; it used to panic on every request.
	writeConfigDir(t, `
setup:
    google-api-key: testkey
    podcast-refresh-interval: 1hr
`)

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for an invalid duration")
	}
	if !strings.Contains(err.Error(), "podcast-refresh-interval") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
	if !strings.Contains(err.Error(), "1h") {
		t.Errorf("error should suggest valid syntax, got: %v", err)
	}
}

func TestLoad_RejectsInvalidEpisodeDurationMinimum(t *testing.T) {
	writeConfigDir(t, `
setup:
    google-api-key: testkey
ytdlp:
    episode-duration-minimum: 5min
`)

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for an invalid duration")
	}
	if !strings.Contains(err.Error(), "episode-duration-minimum") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

func TestLoad_RejectsNegativeDuration(t *testing.T) {
	writeConfigDir(t, `
setup:
    google-api-key: testkey
    podcast-refresh-interval: -5m
`)

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a negative duration")
	}
}

func TestLoad_EnvOverridesProperties(t *testing.T) {
	writeConfigDir(t, `
setup:
    google-api-key: testkey
    podcast-refresh-interval: 30s
`)
	t.Setenv("PODCAST_REFRESH_INTERVAL", "12h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Setup.PodcastRefreshInterval != 12*time.Hour {
		t.Errorf("refresh interval = %v, want 12h", cfg.Setup.PodcastRefreshInterval)
	}
}
