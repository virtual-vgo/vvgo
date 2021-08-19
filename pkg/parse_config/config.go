package parse_config

import "github.com/kelseyhightower/envconfig"

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
		Endpoint string `json:"endpoint" default:"https://discord.com/api/v8"`

		// BotAuthenticationToken is used for making queries about our discord guild.
		// This is found in the bot tab for the discord app.
		BotAuthenticationToken string `json:"bot_authentication_token"`

		// OAuthClientSecret is the secret used in oauth requests.
		// This is found in the oauth2 tab for the discord app.
		OAuthClientSecret string `json:"oauth_client_secret"`
	} `json:"discord" envconfig:"discord"`

	Sheets struct {
		WebsiteDataSpreadsheetID string `json:"website_data_spreadsheet_id" envconfig:"website_data_spreadsheet_id"`
	} `json:"sheets" envconfig:"sheets"`
}

func init() { envconfig.MustProcess("", &Config) }
