package config

import (
	"bytes"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

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
		Address  string `json:"address" envconfig:"address" default:"localhost:6379"`
		UseDB    int    `json:"use_db" envconfig:"USE_DB" default:"0"`
		User     string `json:"user" envconfig:"USER" default:"default"`
		Pass     string `json:"pass" envconfig:"PASS" default:""`
		UseTLS   bool   `json:"use_tls" envconfig:"USE_TLS" default:"false"`
		PoolSize int    `json:"pool_size" envconfig:"POOL_SIZE" default:"10"`
	} `json:"redis" envconfig:"redis"`

	Cloudflare struct {
		ApiKey string `json:"api_key" envconfig:"API_KEY"`
		ZoneId string `json:"zone_id" envconfig:"ZONE_ID"`
	} `json:"cloudflare" envconfig:"CLOUDFLARE"`
}

func init() { ProcessEnv() }

func ProcessEnv() { envconfig.MustProcess("", &Config) }

func ProcessEnvFile(envFile string) {
	defer ProcessEnv()

	file, err := os.Open(envFile)
	if err != nil {
		logrus.WithField("file_name", envFile).WithError(err).Error("os.Open() failed")
		logrus.Fatal("cannot read environment file")
		return
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(file); err != nil {
		logrus.WithField("file_name", envFile).WithError(err).Error("file.Read() failed")
		logrus.Fatal("cannot read environment file")
		return
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "=", 2)
		if len(fields) != 2 {
			logrus.Fatal("cannot parse environment file")
			return
		}

		key, val := fields[0], fields[1]
		if err = os.Setenv(key, val); err != nil {
			logrus.WithField("file_name", envFile).WithError(err).Error("os.Setenv() failed")
			return
		}
	}
}
