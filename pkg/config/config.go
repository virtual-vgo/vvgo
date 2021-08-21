package config

import (
	"bytes"
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"os"
	"strings"
)

var logger = log.New()

var Config struct {
	VVGO struct {
		ListenAddress      string `json:"listen_address" envconfig:"listen_address" default:"0.0.0.0:8080"`
		ServerUrl          string `json:"server_url" envconfig:"server_url" default:"https://vvgo.org"`
		DistroBucket       string `json:"distro_bucket" envconfig:"distro_bucket" default:"vvgo-distro"`
		MemberPasswordHash string `json:"member_password_hash" envconfig:"member_password_hash"`
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

	Sheets struct {
		WebsiteDataSpreadsheetID string `json:"website_data_spreadsheet_id" envconfig:"website_data_spreadsheet_id"`
	} `json:"sheets" envconfig:"sheets"`

	Redis struct {
		Network  string `json:"network" envconfig:"network" default:"tcp"`
		Address  string `json:"address" envconfig:"address" default:"localhost:6379"`
		PoolSize int    `json:"pool_size" envconfig:"pool_size" default:"10"`
	} `json:"redis" envconfig:"redis"`
}

func init() { ProcessEnv() }

func ProcessEnv() { envconfig.MustProcess("", &Config) }

func ProcessEnvFile(envFile string) {
	defer ProcessEnv()

	ctx := context.Background()
	file, err := os.Open(envFile)
	if err != nil {
		logger.WithField("file_name", envFile).MethodFailure(ctx, "os.Open", err)
		logger.Fatal("cannot read environment file")
		return
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(file); err != nil {
		logger.WithField("file_name", envFile).MethodFailure(ctx, "file.Read", err)
		logger.Fatal("cannot read environment file")
		return
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "=", 2)
		if len(fields) != 2 {
			logger.Fatal("cannot parse environment file")
			return
		}

		key, val := fields[0], fields[1]
		if err = os.Setenv(key, val); err != nil {
			logger.WithField("file_name", envFile).MethodFailure(ctx, "os.Setenv", err)
			logger.Fatal("cannot update environment variables")
			return
		}
	}
}
