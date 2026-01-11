package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Setup struct {
		GoogleApiKey           string `mapstructure:"google-api-key" validate:"required"`
		AudioDir               string
		Cron                   string `mapstructure:"cron"`
		ConfigDir              string `mapstructure:"config-dir" validate:"required"`
		DbFile                 string
		PodcastRefreshInterval string `mapstructure:"podcast-refresh-interval"`
	} `mapstructure:"setup"`

	Ntfy struct {
		Server         string `mapstructure:"server"`
		Topic          string `mapstructure:"topic"`
		Authentication struct {
			Token     string `mapstructure:"token"`
			BasicAuth struct {
				Username string `mapstructure:"username" validate:"required_with=password"`
				Password string `mapstructure:"password" validate:"required_with=username"`
			} `mapstructure:"basic-auth"`
		} `mapstructure:"authentication"`
	} `mapstructure:"ntfy"`

	Authentication struct {
		Token     string `mapstructure:"token"`
		BasicAuth struct {
			Username string `mapstructure:"username" validate:"required_with=password"`
			Password string `mapstructure:"password" validate:"required_with=username"`
		} `mapstructure:"basic-auth"`
	} `mapstructure:"authentication"`

	Ytdlp struct {
		CookiesFile            string `mapstructure:"cookies-file"`
		SponsorBlockCategories string `mapstructure:"sponsorblock-categories"`
		EpisodeDurationMinimum string `mapstructure:"episode-duration-minimum"`
		YtdlpExtractorArgs     string `mapstructure:"ytdlp-extractor-args"`
	} `mapstructure:"ytdlp"`
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

	v.SetDefault("ytdlp.episode-duration-minimum", "3m")
	v.SetDefault("setup.podcast-refresh-interval", "1h")
	v.SetDefault("setup.config-dir", configDir)
	v.SetDefault("setup.audio-dir", "audio")
	v.SetDefault("ytdlp.sponsorblock-categories", "sponsor")
	v.SetDefault("setup.google-api-key", os.Getenv("GOOGLE_API_KEY"))

	replacer := strings.NewReplacer(".", "_", "-", "_")
	v.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("env")
	v.AutomaticEnv()

	v.BindEnv("setup.config-dir", "CONFIG_DIR")
	v.BindEnv("setup.audio-dir", "AUDIO_DIR")
	v.BindEnv("setup.google-api-key", "GOOGLE_API_KEY")
	v.BindEnv("ytdlp.cookies-file", "COOKIES_FILE")
	v.BindEnv("ntfy.server", "NTFY_SERVER")
	v.BindEnv("ntfy.topic", "NTFY_TOPIC")
	v.BindEnv("authentication.token", "TOKEN")
	v.BindEnv("setup.cron", "CRON")
	v.BindEnv("ytdlp.episode-duration-minimum", "MIN_DURATION")
	v.BindEnv("ytdlp.sponsorblock-categories", "SPONSORBLOCK_CATEGORIES")
	v.BindEnv("ytdlp.ytdlp-extractor-args", "YTDLP_EXTRACTOR_ARGS")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
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
		Config.MinDuration = "3m"
	}
}
