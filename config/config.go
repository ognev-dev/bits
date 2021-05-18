package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DevEnv     = "DEV"
	TestEnv    = "TEST"
	ReleaseEnv = "RELEASE"
)

type Config struct {
	AppEnv string `yaml:"app_env"`
	DB     struct {
		Addr       string
		User       string
		Password   string
		Name       string
		LogQueries bool `yaml:"log_queries"`
	}

	Server struct {
		Port        string
		Host        string
		ApiBasePath string `yaml:"api_base_path"`

		// Static directive defines local and web paths.
		// Anything that is requested from "WebPath" will served from "LocalPath" as-is
		// For example if local path is set to "./web" and web path is set to "/static/"
		// requesting "/static/something.html" will serve "./web/something.html" if exists
		Static struct {
			LocalPath string `yaml:"local_path"`
			WebPath   string `yaml:"web_path"`
		}
	}

	Dir struct {
		Data string
		Logs string
	}

	GitHub struct {
		DataRepo             string `yaml:"data_repo"`
		DataPushedHookSecret string `yaml:"data_pushed_hook_secret"`

		// Data Warden is GitHub app that helps with data checks and validation
		// https://github.com/apps/data-warden
		DataWarden struct {
			AppID     int64  `yaml:"app_id"`     // GitHub App ID
			InstallID int64  `yaml:"install_id"` // Installation ID
			PEMPath   string `yaml:"pem_path"`   // path to private-key.pem
		} `yaml:"data_warden"`

		// https://github.com/settings/applications/new
		// https://docs.github.com/en/free-pro-team@latest/developers/apps/authorizing-oauth-apps
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"github"`

	AuthTokenLifetime time.Duration `yaml:"auth_token_life_time"`
}

var conf *Config

// Get returns config from .config.yaml
func Get() *Config {
	if conf != nil {
		return conf
	}

	name := ".config.yaml"
	yamlConf, err := ioutil.ReadFile(name)
	if err != nil {
		msg := ".config.yaml missing from working directory"
		wd, err := os.Getwd()
		if err == nil {
			msg += fmt.Sprintf(" (%s)", wd)
		}
		panic(msg)
	}
	err = yaml.Unmarshal(yamlConf, &conf)
	if err != nil {
		panic(err)
	}

	return conf
}

func Reload() *Config {
	conf = nil
	return Get()
}

func IsDevEnv() bool {
	return conf.AppEnv == DevEnv
}

func IsTestEnv() bool {
	return conf.AppEnv == TestEnv
}

func IsReleaseEnv() bool {
	return conf.AppEnv == ReleaseEnv
}
