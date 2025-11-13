package config

import (
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"os"
	"path"
)

type ConfigDef struct {
	ConfigDir              string
	AudioDir               string
	DbFile                 string
	CookiesFile            string
	Token                  string
	GoogleApiKey           string
	SponsorBlockCategories string
	Cron                   string
	MinDuration            string
	AudioFormat            string
	AudioQuality           string
	BackupCron             string
	BackupIncludeAudio     bool
}

var Config *ConfigDef

func init() {
	Config = &ConfigDef{}

	Config.ConfigDir = os.Getenv("CONFIG_DIR")
	if Config.ConfigDir == "" {
		Config.ConfigDir = "/config"
	}
	Config.AudioDir = path.Join(Config.ConfigDir, "audio")
	Config.DbFile = path.Join(Config.ConfigDir, "sqlite.db")

	cookiesFile := os.Getenv("COOKIES_FILE")
	if cookiesFile != "" {
		Config.CookiesFile = path.Join(Config.ConfigDir, cookiesFile)
		println("CONFIG | Cookies file set: " + Config.CookiesFile)
	}

	Config.Token = os.Getenv("TOKEN")
	if Config.Token != "" {
		println("CONFIG | Token set.")
	}

	Config.GoogleApiKey = os.Getenv("GOOGLE_API_KEY")
	if Config.GoogleApiKey == "" {
		panic("GOOGLE_API_KEY is not set")
	}

	Config.SponsorBlockCategories = os.Getenv("SPONSORBLOCK_CATEGORIES")
	if Config.SponsorBlockCategories != "" {
		println("CONFIG | Sponsor block segments defined: " + Config.SponsorBlockCategories)
	}

	Config.Cron = os.Getenv("CRON")
	if Config.Cron != "" {
		println("CONFIG | Cron set: " + Config.Cron)
	}

	Config.MinDuration = os.Getenv("MIN_DURATION")
	if Config.MinDuration == "" {
		Config.MinDuration = constants.DefaultMinDuration
	}

	Config.AudioFormat = os.Getenv("AUDIO_FORMAT")
	if Config.AudioFormat == "" {
		Config.AudioFormat = "m4a"
	}
	println("CONFIG | Audio format set: " + Config.AudioFormat)

	Config.AudioQuality = os.Getenv("AUDIO_QUALITY")
	if Config.AudioQuality == "" {
		Config.AudioQuality = "192k"
	}
	println("CONFIG | Audio quality set: " + Config.AudioQuality)

	Config.BackupCron = os.Getenv("BACKUP_CRON")
	if Config.BackupCron != "" {
		println("CONFIG | Backup cron set: " + Config.BackupCron)
	}

	Config.BackupIncludeAudio = os.Getenv("BACKUP_INCLUDE_AUDIO") == "true"
	if Config.BackupIncludeAudio {
		println("CONFIG | Backup will include audio files")
	}
}
