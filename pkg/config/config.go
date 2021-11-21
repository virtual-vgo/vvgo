package config

import (
	"bytes"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const EnvPrefix=""

var Config struct {
	Development bool

	VVGO struct {
		ListenAddress      string `json:"listen_address" envconfig:"listen_address" default:"0.0.0.0:8080"`
		ServerUrl          string `json:"server_url" envconfig:"server_url" default:"https://vvgo.org"`
		DistroBucket       string `json:"distro_bucket" envconfig:"distro_bucket" default:"vvgo-distro"`
		MemberPasswordHash string `json:"member_password_hash" envconfig:"member_password_hash"`
		ClientToken        string `json:"vvgo_client_token" envconfig:"vvgo_client_token"`
	} `json:"vvgo" envconfig:"vvgo"`

	Minio struct {
		Endpoint  string `json:"endpoint" envconfig:"endpoint" default:"localhost:9000"`
		AccessKey string `json:"access_key" envconfig:"access_key" default:"minioadmin"`
		SecretKey string `json:"secret_key" envconfig:"secret_key" default:"minioadmin"`
		UseSSL    bool   `json:"use_ssl" envconfig:"use_ssl" default:"false"`
	} `json:"minio" envconfig:"minio"`

	Discord struct {
		// Endpoint is the api endpoint to query. Defaults to https://discord.com/api/v8.
		// This should only be overwritten for testing.
		Endpoint string `json:"endpoint" envconfig:"endpoint" default:"https://discord.com/api/v8"`

		// BotAuthenticationToken is used for making queries about our discord guild.
		// This is found in the bot tab for the discord app.
		BotAuthenticationToken string `json:"bot_authentication_token" envconfig:"bot_authentication_token"`

		// OAuthClientSecret is the secret used in oauth requests.
		// This is found in the oauth2 tab for the discord app.
		OAuthClientSecret string `json:"oauth_client_secret" envconfig:"oauth_client_secret"`
	} `json:"discord" envconfig:"discord"`

	Redis struct {
		Network  string `json:"network" envconfig:"network" default:"tcp"`
		Address  string `json:"address" envconfig:"address" default:"localhost:6379"`
		PoolSize int    `json:"pool_size" envconfig:"pool_size" default:"10"`
	} `json:"redis" envconfig:"redis"`
}

func init() { ProcessEnv() }

func ProcessEnv() { envconfig.MustProcess(EnvPrefix, &Config) }

func ProcessEnvFile(envFile string) {
	defer ProcessEnv()

	if envFile == "" {
		return
	}

	file, err := os.Open(envFile)
	if os.IsNotExist(err) {
		logrus.WithField("file_name", envFile).Infof("env file does not exist, skipping")
		return
	} else if err != nil {
		logrus.WithField("file_name", envFile).WithError(err).Error("os.Open() failed")
		logrus.Fatal("cannot read environment file")
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(file); err != nil {
		logrus.WithField("file_name", envFile).WithError(err).Error("file.Read() failed")
		logrus.Fatal("cannot read environment file")
		return
	}

	for i, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "=", 2)
		if len(fields) != 2 {
			logrus.WithField("file_name", envFile).
				WithField("line", i).
				WithField("text", line).
				Error("cannot parse line, skipping")
			return
		}

		key, val := fields[0], fields[1]
		if os.Getenv(key) != "" {
			continue
		}
		if err = os.Setenv(key, val); err != nil {
			logrus.WithField("file_name", envFile).WithError(err).Error("os.Setenv() failed")
			return
		}
	}
}
