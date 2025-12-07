package config

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	gotenv "github.com/subosito/gotenv"
)

type Config struct {
	Setup struct {
		GoogleApiKey string `mapstructure:"google-api-key" validate:"required"`
		AudioDir     string
		Cron         string `mapstructure:"cron"`
		ConfigDir    string `mapstructure:"config-dir" validate:"required"`
		DbFile       string
	} `mapstructure:"setup"`

	Ntfy struct {
		Server string `mapstructure:"server"`
		Topic  string `mapstructure:"topic"`
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

var validate = validator.New()

var AppConfig *Config

func init() {
	_ = gotenv.Load()
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "/config"
	}
	audioDir := os.Getenv("AUDIO_DIR")
	if audioDir == "" {
		audioDir = path.Join(configDir, "audio")
	}

	v := viper.New()
	v.SetConfigName("properties")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	v.SetDefault("ytdlp.episode-duration-minimum", "3m")
	v.SetDefault("setup.config-dir", configDir)
	v.SetDefault("setup.audio-dir", "audio")
	v.SetDefault("ytdlp.sponsorblock-categories", "sponsor")

	v.ReadInConfig()

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
		log.Fatalf("unmarshal config: %w", err)
	}

	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("invalid config: %w", err)
	}

	AppConfig = &cfg
	if AppConfig.Ytdlp.CookiesFile != "" {
		AppConfig.Ytdlp.CookiesFile = path.Join(AppConfig.Setup.ConfigDir, AppConfig.Ytdlp.CookiesFile)
	}
	AppConfig.Setup.DbFile = path.Join(AppConfig.Setup.ConfigDir, "sqlite.db")
	AppConfig.Setup.AudioDir = audioDir
}
