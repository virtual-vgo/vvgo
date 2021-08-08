package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var logger = log.New()
var ErrNon200Response = errors.New("non-200 response from discord")
var ErrInvalidOAuthCode = errors.New("invalid oauth code")

const ClientPublicKey = "a56a084a21829d02f272e4e3f4b67a846a831281849f6740f7bbf873840c4076"
const ApplicationID = "700963768787795998"
const OAuthClientID = ApplicationID
const VVGOGuildID = "690626216637497425" // The VVGO discord server
const VVGOVerifiedMemberRoleID = "690636730281230396"
const VVGOProductionTeamRoleID = "746434659252174971"
const VVGOExecutiveDirectorRoleID = "690626333062987866"
const VVGOProductionDirectorRoleID = "805504313072943155"

// Client that makes discord requests.
type Client struct{}

// Config for discord requests.
type Config struct {
	// Api endpoint to query. Defaults to https://discord.com/api/v8.
	// This should only be overwritten for testing.
	Endpoint string `json:"endpoint" default:"https://discord.com/api/v8"`

	// BotAuthenticationToken is used for making queries about our discord guild.
	// This is found in the bot tab for the discord app.
	BotAuthenticationToken string `json:"bot_authentication_token"`

	// OAuthClientSecret is the secret used in oauth requests.
	// This is found in the oauth2 tab for the discord app.
	OAuthClientSecret string `json:"oauth_client_secret"`

	// OAuthRedirectURI is the redirect uri we set in discord.
	// This is found in the oauth2 tab for the discord app.
	OAuthRedirectURI string `json:"oauth_redirect_uri"`
}

func LoginURL(ctx context.Context, state string) string {
	var config Config
	ctx = parse_config.ReadModuleConfig(ctx, "discord", &config)

	query := make(url.Values)
	query.Set("client_id", OAuthClientID)
	query.Set("redirect_uri", config.OAuthRedirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("scope", "identify")
	return "https://discord.com/api/oauth2/authorize?" + query.Encode()
}

// Query oauth token from discord.
// We use the authorization code grant.
func QueryOAuth(ctx context.Context, code string) (*OAuthToken, error) {
	req, err := newOAuthRequest(ctx, code)
	if err != nil {
		return nil, err
	}

	var oauthToken OAuthToken
	if _, err := doDiscordRequest(req, &oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

const ConfigModule = "discord"

func readConfig(ctx context.Context) Config {
	var config Config
	ctx = parse_config.ReadModuleConfig(ctx, ConfigModule, &config)
	return config
}

// build the oauth request
func newOAuthRequest(ctx context.Context, code string) (*http.Request, error) {
	var config Config
	ctx = parse_config.ReadModuleConfig(ctx, ConfigModule, &config)

	if code == "" {
		return nil, ErrInvalidOAuthCode
	}

	// build the authorization request
	form := make(url.Values)
	form.Add("client_id", OAuthClientID)
	form.Add("client_secret", config.OAuthClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", config.OAuthRedirectURI)
	form.Add("scope", "identify")

	req, err := newRequest(ctx, http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// Query discord for the token's identity.
// This requires an oauth token with identity scope.
// https://discordapp.com/developers/docs/resources/user#get-current-user
func QueryIdentity(ctx context.Context, oauthToken *OAuthToken) (*User, error) {
	req, err := newTokenRequest(ctx, oauthToken, "/users/@me")
	if err != nil {
		return nil, err
	}

	var discordUser User
	if _, err := doDiscordRequest(req, &discordUser); err != nil {
		return nil, err
	}
	return &discordUser, nil
}

// returns a request using an oauth token for authentication
func newTokenRequest(ctx context.Context, oauthToken *OAuthToken, path string) (*http.Request, error) {
	// build the request
	req, err := newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.AccessToken))
	return req, err
}

// Query discord for the guild member object of the guild id and user id.
// Here we use the the server's own auth token.
// https://discordapp.com/developers/docs/resources/guild#get-guild-member
func QueryGuildMember(ctx context.Context, userID Snowflake) (*GuildMember, error) {
	req, err := newBotRequest(ctx, http.MethodGet, "/guilds/"+VVGOGuildID+"/members/"+userID.String(), nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		return nil, err
	}

	// unmarshal the response
	var guildMember GuildMember
	if _, err := doDiscordRequest(req, &guildMember); err != nil {
		return nil, err
	}
	return &guildMember, nil
}

func GetApplicationCommands(ctx context.Context) ([]ApplicationCommand, error) {
	req, err := newSlashCommandRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var commands []ApplicationCommand
	_, err = doDiscordRequest(req, &commands)
	return commands, err
}

func CreateApplicationCommand(ctx context.Context, params CreateApplicationCommandParams) (*ApplicationCommand, error) {
	var paramsBytes bytes.Buffer
	if err := json.NewEncoder(&paramsBytes).Encode(params); err != nil {
		return nil, fmt.Errorf("json.Encode() failed: %w", err)
	}

	req, err := newSlashCommandRequest(ctx, http.MethodPost, &paramsBytes)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var command ApplicationCommand
	_, err = doDiscordRequest(req, &command)
	if err != nil {
		return nil, err
	}
	return &command, nil
}

func DeleteApplicationCommand(ctx context.Context, id Snowflake) error {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands/" + id.String()
	req, err := newBotRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	_, err = doDiscordRequest(req, nil)
	return err
}

func CreateMessage(ctx context.Context, channelId Snowflake, params CreateMessageParams) (*Message, error) {
	path := "/channels/" + channelId.String() + "/messages"

	var message Message
	err := doDiscordBotRequestWithJsonParams(ctx, path, http.MethodPost, &params, &message)
	return &message, err
}

func EditMessage(ctx context.Context, channelId Snowflake, messageId Snowflake, params EditMessageParams) (*Message, error) {
	path := "/channels/" + channelId.String() + "/messages/" + messageId.String()

	var message Message
	err := doDiscordBotRequestWithJsonParams(ctx, path, http.MethodPatch, &params, &message)
	return &message, err
}

func BulkDeleteMessages(ctx context.Context, channelId Snowflake, params BulkDeleteMessagesParams) error {
	path := "/channels/" + channelId.String() + "/messages/bulk-delete"
	return doDiscordBotRequestWithJsonParams(ctx, path, http.MethodPost, &params, nil)
}

func newSlashCommandRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands"
	return newBotRequest(ctx, method, path, body)
}

func doDiscordBotRequestWithJsonParams(ctx context.Context, path, method string, params interface{}, dest interface{}) error {
	var paramsBytes bytes.Buffer
	if err := json.NewEncoder(&paramsBytes).Encode(params); err != nil {
		return fmt.Errorf("json.Encode() failed: %w", err)
	}

	req, err := newBotRequest(ctx, method, path, &paramsBytes)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = doDiscordRequest(req, dest)
	return err
}

// returns a request using a bot token for authentication
func newBotRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	var config Config
	ctx = parse_config.ReadModuleConfig(ctx, "discord", &config)
	req, err := newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+config.BotAuthenticationToken)
	return req, err
}

func newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	var config Config
	ctx = parse_config.ReadModuleConfig(ctx, "discord", &config)

	req, err := http.NewRequestWithContext(ctx, method, config.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %w", err)
	}
	return req, nil
}

// performs the http request and logs results
func doDiscordRequest(req *http.Request, dest interface{}) (*http.Response, error) {
	resp, err := http_wrappers.DoRequest(req)
	switch {
	case err != nil:
		logger.WithError(err).Error("http.Do() failed")
		return nil, err

	case resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent:
		err = ErrNon200Response
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)

		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
			"url":    req.URL.String(),
			"body":   buf.String(),
		}).Error("non-200 response from discord")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"method": req.Method,
		"status": resp.StatusCode,
		"url":    req.URL.String(),
	}).Info("discord api request complete")
	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(dest)
	}
	return resp, nil
}
