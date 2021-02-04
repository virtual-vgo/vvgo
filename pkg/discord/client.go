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

var logger = log.Logger()
var ErrNon200Response = errors.New("non-200 response from discord")
var ErrInvalidOAuthCode = errors.New("invalid oauth code")

const ClientPublicKey = "a56a084a21829d02f272e4e3f4b67a846a831281849f6740f7bbf873840c4076"
const ApplicationID = "700963768787795998"
const OAuthClientID = ApplicationID
const VVGOGuildID  = "690626216637497425" // The VVGO discord server
const VVGOVerifiedMemberRoleID  = "690636730281230396"
const VVGOProductionTeamRoleID  = "746434659252174971"
const VVGOExecutiveDirectorRoleID  = "690626333062987866"

// Client that makes discord requests.
type Client struct {
	Config Config
}

func NewClient(ctx context.Context) *Client {
	return &Client{Config: newConfig(ctx)}
}

// Config for discord requests.
type Config struct {
	// Api endpoint to query. Defaults to https://discord.com/api/v8.
	// This should only be overwritten for testing.
	Endpoint string `redis:"endpoint" default:"https://discord.com/api/v8"`

	// BotAuthenticationToken is used for making queries about our discord guild.
	// This is found in the bot tab for the discord app.
	BotAuthenticationToken string `redis:"bot_authentication_token"`

	// OAuthClientSecret is the secret used in oauth requests.
	// This is found in the oauth2 tab for the discord app.
	OAuthClientSecret string `redis:"oauth_client_secret"`

	// OAuthRedirectURI is the redirect uri we set in discord.
	// This is found in the oauth2 tab for the discord app.
	OAuthRedirectURI string `redis:"oauth_redirect_uri"`
}

func newConfig(ctx context.Context) Config {
	var dest Config
	parse_config.SetDefaults(&dest)
	if err := parse_config.ReadFromRedisHash(ctx, "discord", &dest); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
	}
	return dest
}

func (x Client) LoginURL(state string) string {
	query := make(url.Values)
	query.Set("client_id", OAuthClientID)
	query.Set("redirect_uri", x.Config.OAuthRedirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("scope", "identify")
	return "https://discord.com/api/oauth2/authorize?" + query.Encode()
}

// Query oauth token from discord.
// We use the authorization code grant.
func (x Client) QueryOAuth(ctx context.Context, code string) (*OAuthToken, error) {
	req, err := x.newOAuthRequest(ctx, code)
	if err != nil {
		return nil, err
	}

	var oauthToken OAuthToken
	if _, err := doDiscordRequest(req, &oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// build the oauth request
func (x Client) newOAuthRequest(ctx context.Context, code string) (*http.Request, error) {
	if code == "" {
		return nil, ErrInvalidOAuthCode
	}

	// build the authorization request
	form := make(url.Values)
	form.Add("client_id", OAuthClientID)
	form.Add("client_secret", x.Config.OAuthClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", x.Config.OAuthRedirectURI)
	form.Add("scope", "identify")

	req, err := x.newRequest(ctx, http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Query discord for the token's identity.
// This requires an oauth token with identity scope.
// https://discordapp.com/developers/docs/resources/user#get-current-user
func (x Client) QueryIdentity(ctx context.Context, oauthToken *OAuthToken) (*User, error) {
	req, err := x.newTokenRequest(ctx, oauthToken, "/users/@me")
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
func (x Client) newTokenRequest(ctx context.Context, oauthToken *OAuthToken, path string) (*http.Request, error) {
	// build the request
	req, err := x.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.AccessToken))
	return req, err
}

// Query discord for the guild member object of the guild id and user id.
// Here we use the the server's own auth token.
// https://discordapp.com/developers/docs/resources/guild#get-guild-member
func (x Client) QueryGuildMember(ctx context.Context, userID Snowflake) (*GuildMember, error) {
	req, err := x.newBotRequest(ctx, http.MethodGet, "/guilds/"+VVGOGuildID+"/members/"+userID.String(), nil)
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

func (x Client) GetApplicationCommands(ctx context.Context) ([]ApplicationCommand, error) {
	req, err := x.newSlashCommandRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var commands []ApplicationCommand
	_, err = doDiscordRequest(req, &commands)
	return commands, err
}

func (x Client) CreateApplicationCommand(ctx context.Context, params CreateApplicationCommandParams) (*ApplicationCommand, error) {
	var paramsBytes bytes.Buffer
	if err := json.NewEncoder(&paramsBytes).Encode(params); err != nil {
		return nil, fmt.Errorf("json.Encode() failed: %w", err)
	}

	req, err := x.newSlashCommandRequest(ctx, http.MethodPost, &paramsBytes)
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

func (x Client) DeleteApplicationCommand(ctx context.Context, id Snowflake) error {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands/" + id.String()
	req, err := x.newBotRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	_, err = doDiscordRequest(req, nil)
	return err
}

func (x Client) newSlashCommandRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands"
	return x.newBotRequest(ctx, method, path, body)
}

// returns a request using a bot token for authentication
func (x Client) newBotRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := x.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+x.Config.BotAuthenticationToken)
	return req, err
}

func (x Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	endpoint := x.Config.Endpoint
	req, err := http.NewRequestWithContext(ctx, method, endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %w", err)
	}
	return req, err
}

// performs the http request and logs results
func doDiscordRequest(req *http.Request, dest interface{}) (resp *http.Response, err error) {
	resp, err = http_wrappers.DoRequest(req)
	switch {
	case err != nil:
		logger.WithError(err).Error("http.Do() failed")

	case resp.StatusCode != http.StatusOK:
		err = ErrNon200Response
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)

		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
			"url":    req.URL.String(),
			"body":   buf.String(),
		}).Error("non-200 response from discord")

	default:
		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
			"url":    req.URL.String(),
		}).Info("discord api request complete")
		if dest != nil {
			err = json.NewDecoder(resp.Body).Decode(dest)
		}
	}
	return
}
