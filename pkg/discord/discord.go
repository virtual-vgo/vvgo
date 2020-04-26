package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var logger = log.Logger()

// Client that makes discord requests.
type Client struct {
	config Config
}

func NewClient(config Config) *Client {
	return &Client{config: config}
}

var ErrInvalidOAuthCode = errors.New("invalid oauth code")

// Config for discord requests.
type Config struct {
	// Api endpoint to query
	Endpoint string `default:"https://discordapp.com/api/v6"`

	// BotAuthToken is used for making queries about our discord guild.
	// This is found in the bot tab for the discord app.
	BotAuthToken string `split_words:"true"`

	// OAuthClientID is the client id used in oauth requests.
	// This is found in the oauth2 tab for the discord app.
	OAuthClientID string `envconfig:"oauth_client_id"`

	// OAuthClientSecret is the secret used in oauth requets.
	// This is found in the oauth2 tab for the discord app.
	OAuthClientSecret string `envconfig:"oauth_client_secret"`

	// OAuthRedirectURI is the redirect uri we set in discord.
	// This is found in the oauth2 tab for the discord app.
	OAuthRedirectURI string `envconfig:"oauth_redirect_uri"`
}

// This is the oauth token returned by discord after a successful oauth request.
// https://discordapp.com/developers/docs/topics/oauth2
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// A discord user id.
// These are string encoded 64bit ints.
// https://discordapp.com/developers/docs/reference#snowflakes
type UserID string

func (x UserID) String() string { return string(x) }

// A discord guild id.
// These are string encoded 64bit ints.
// https://discordapp.com/developers/docs/reference#snowflakes
type GuildID string

func (x GuildID) String() string { return string(x) }

// A discord user object.
// We only care about the id.
// https://discordapp.com/developers/docs/resources/user#user-object
type User struct {
	ID UserID `json:"id"`
}

// A discord guild member object.
// We only care about the roles.
// https://discordapp.com/developers/docs/resources/guild#guild-member-object
type GuildMember struct {
	Roles []string `json:"roles"`
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
	form.Add("client_id", x.config.OAuthClientID)
	form.Add("client_secret", x.config.OAuthClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", x.config.OAuthRedirectURI)
	form.Add("scope", "identify")

	req, err := x.newRequest(ctx, http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %v", err)
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
func (x Client) QueryGuildMember(ctx context.Context, guildID GuildID, userID UserID) (*GuildMember, error) {
	req, err := x.newBotRequest(ctx, x.config.BotAuthToken, "/guilds/"+guildID.String()+"/members/"+userID.String())
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

// returns a request using a bot token for authentication
func (x Client) newBotRequest(ctx context.Context, token string, path string) (*http.Request, error) {
	req, err := x.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+token)
	return req, err
}

func (x Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, x.config.Endpoint+path, body)
}

// performs the http request and logs results
func doDiscordRequest(req *http.Request, dest interface{}) (*http.Response, error) {
	resp, err := tracing.DoHttpRequest(req)
	switch {
	case err != nil:
		logger.WithError(err).Error("tracing.DoHttpRequest() failed")

	case resp.StatusCode != http.StatusOK:
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
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
		err = json.NewDecoder(resp.Body).Decode(dest)
		if err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
	}
	return resp, err
}
