package config

import (
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
)

type ConfigDef struct {
	ConfigDir              string `yaml:"-"`
	AudioDir               string `yaml:"audio-dir"`
	DbFile                 string `yaml:"-"`
	CookiesFile            string `yaml:"cookies-file"`
	GoogleApiKey           string `yaml:"google-api-key"`
	SponsorBlockCategories string `yaml:"sponsorblock-categories"`
	Cron                   string `yaml:"episode-delete-cron"`
	MinDuration            string `yaml:"min-duration"`
	YtdlpExtractorArgs     string `yaml:"ytdlp-extractor-args"`
}

var Config *ConfigDef
var Ntfy *NtfyDef
var Authentication *AuthenticationDef

type NtfyDef struct {
	Server string `yaml:"server"`
	Topic  string `yaml:"topic"`
}

type AuthenticationDef struct {
	Token     string    `yaml:"token"`
	BasicAuth BasicAuth `yaml:"basic-auth"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type propertiesFileStruct struct {
	Ntfy           NtfyDef           `yaml:"ntfy"`
	Authentication AuthenticationDef `yaml:"authentication"`
	Setup          ConfigDef         `yaml:"setup"`
}

var propertiesConfig propertiesFileStruct

func init() {
	Config = &ConfigDef{}
	Ntfy = &NtfyDef{}
	Authentication = &AuthenticationDef{}

	Config.ConfigDir = os.Getenv("CONFIG_DIR")
	if Config.ConfigDir == "" {
		Config.ConfigDir = "/config"
	}

	Config.DbFile = path.Join(Config.ConfigDir, "sqlite.db")

	setupPropertiesFile()
	setProperties()
}

func setProperties() {
	setupAudioDir()
	setupCookiesFile()
	setupSecurity()
	setupGoogleApiKey()
	setupYtdlp()
	setupAdditionalProps()
	setupNtfy()
}

func setupNtfy() {
	if s := os.Getenv("NTFY_SERVER"); s != "" {
		Ntfy.Server = s
		println("CONFIG | Ntfy server set from env.")
	} else if propertiesConfig.Ntfy.Server != "" {
		Ntfy.Server = propertiesConfig.Ntfy.Server
		println("CONFIG | Ntfy server set from properties.yml.")
	}

	if t := os.Getenv("NTFY_TOPIC"); t != "" {
		Ntfy.Topic = t
		println("CONFIG | Ntfy topic set from env.")
	} else if propertiesConfig.Ntfy.Topic != "" {
		Ntfy.Topic = propertiesConfig.Ntfy.Topic
		println("CONFIG | Ntfy topic set from properties.yml.")
	}
}

func setupAdditionalProps() {
	if c := os.Getenv("CRON"); c != "" {
		Config.Cron = c
		println("CONFIG | Cron set from env: " + Config.Cron)
	} else if propertiesConfig.Setup.Cron != "" {
		Config.Cron = propertiesConfig.Setup.Cron
		println("CONFIG | Cron set from properties.yml: " + Config.Cron)
	}

	if m := os.Getenv("MIN_DURATION"); m != "" {
		Config.MinDuration = m
		println("CONFIG | MinDuration set (from env): " + Config.MinDuration)
	} else if propertiesConfig.Setup.MinDuration != "" {
		Config.MinDuration = propertiesConfig.Setup.MinDuration
		println("CONFIG | MinDuration set from properties.yml: " + Config.MinDuration)
	} else {
		Config.MinDuration = "3m"
	}
}

func setupYtdlp() {
	if s := os.Getenv("SPONSORBLOCK_CATEGORIES"); s != "" {
		Config.SponsorBlockCategories = s
		println("CONFIG | Sponsor block segments defined (from env).")
	} else if propertiesConfig.Setup.SponsorBlockCategories != "" {
		Config.SponsorBlockCategories = propertiesConfig.Setup.SponsorBlockCategories
		println("CONFIG | Sponsor block segments defined from properties.yml.")
	}

	if propertiesConfig.Setup.YtdlpExtractorArgs != "" {
		Config.YtdlpExtractorArgs = propertiesConfig.Setup.YtdlpExtractorArgs
		println("CONFIG | Ytdlp extractor args set from properties.yml.")
	}
}

func setupGoogleApiKey() {
	if g := os.Getenv("GOOGLE_API_KEY"); g != "" {
		Config.GoogleApiKey = g
		println("CONFIG | Google API key set from env.")
	} else if propertiesConfig.Setup.GoogleApiKey != "" {
		Config.GoogleApiKey = propertiesConfig.Setup.GoogleApiKey
		println("CONFIG | Google API key set from properties.yml.")
	}
	if Config.GoogleApiKey == "" {
		panic("GOOGLE_API_KEY is not set")
	}
}

func setupSecurity() {
	if tok := os.Getenv("TOKEN"); tok != "" {
		Authentication.Token = tok
		println("CONFIG | Token set from env.")
	} else if propertiesConfig.Authentication.Token != "" {
		Authentication.Token = propertiesConfig.Authentication.Token
		println("CONFIG | Token set from properties.yml.")
	}

	if propertiesConfig.Authentication.BasicAuth.Username != "" {
		Authentication.BasicAuth.Username = propertiesConfig.Authentication.BasicAuth.Username
	}
	if propertiesConfig.Authentication.BasicAuth.Password != "" {
		Authentication.BasicAuth.Password = propertiesConfig.Authentication.BasicAuth.Password
	}
}

func setupCookiesFile() {
	if cookiesFile := os.Getenv("COOKIES_FILE"); cookiesFile != "" {
		Config.CookiesFile = path.Join(Config.ConfigDir, cookiesFile)
		println("CONFIG | Cookies file set from env.")
	} else if propertiesConfig.Setup.CookiesFile != "" {
		Config.CookiesFile = path.Join(Config.ConfigDir, propertiesConfig.Setup.CookiesFile)
		println("CONFIG | Cookies file set from properties.yml.")
	}
}

func setupAudioDir() {
	if propertiesConfig.Setup.AudioDir != "" {
		Config.AudioDir = propertiesConfig.Setup.AudioDir
	} else {
		Config.AudioDir = path.Join(Config.ConfigDir, "audio")
	}
}

func setupPropertiesFile() {
	configPath := filepath.Join(Config.ConfigDir, "properties.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		emptyConfig := propertiesFileStruct{
			Setup: ConfigDef{
				ConfigDir:              "",
				AudioDir:               "",
				CookiesFile:            "",
				GoogleApiKey:           "",
				SponsorBlockCategories: "",
				Cron:                   "",
				MinDuration:            "",
				YtdlpExtractorArgs:     "",
			},
			Ntfy: NtfyDef{
				Server: "",
				Topic:  "",
			},
			Authentication: AuthenticationDef{
				Token: "",
				BasicAuth: BasicAuth{
					Username: "",
					Password: "",
				},
			},
		}
		data, err := yaml.Marshal(emptyConfig)
		if err != nil {
			log.Error("Failed to marshal empty config:", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			log.Error("Failed to create empty properties.yml:", err)
		}
		propertiesConfig = emptyConfig
		return
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("Failed to read properties.yml:", err)
		return
	}
	if err := yaml.Unmarshal(data, &propertiesConfig); err != nil {
		log.Error("Failed to unmarshal properties.yml:", err)
		return
	}
}
