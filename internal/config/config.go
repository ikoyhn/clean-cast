package config

import (
	"os"
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
}

var Config *ConfigDef

func init() {
	Config = &ConfigDef{}

	Config.ConfigDir = os.Getenv("CONFIG_DIR")
	if Config.ConfigDir == "" {
		Config.ConfigDir = "/config"
	}
	Config.AudioDir = Config.ConfigDir + "/audio/"
	Config.DbFile = Config.ConfigDir + "/sqlite.db"

	cookiesFile := os.Getenv("COOKIES_FILE")
	Config.CookiesFile = Config.ConfigDir + "/" + cookiesFile
	if cookiesFile != "" {
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
}
