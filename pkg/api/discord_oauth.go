package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/url"
	"strings"
)

type DiscordOAuthHandlerConfig struct {
	Endpoint          string `default:"https://discordapp.com/api/v6"`
	BotAuthToken      string `split_words:"true"`
	GuildID           string `split_words:"true"`
	RoleVVGOMember    string `envconfig:"role_vvgo_member"`
	OAuthClientID     string `envconfig:"oauth_client_id"`     // find in discord
	OAuthClientSecret string `envconfig:"oauth_client_secret"` // find in discord
	OAuthRedirectURI  string `envconfig:"oauth_redirect_uri"`  // this is the redirect we set in discord
}

type DiscordOAuthHandler struct {
	Config   DiscordOAuthHandlerConfig
	Sessions *sessions.Store
}

var ErrInvalidOAuthCode = errors.New("invalid oauth code")

type oauthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type discordUser struct {
	ID string `json:"id"`
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordOAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "discord_oauth_handler")
	defer span.Send()

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("httpClient.Do() failed")
			tracing.AddError(ctx, err)
			logger.Error("oauth authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	code := r.FormValue("code")
	oauthToken, err := x.queryDiscordOauth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	discordUser, err := x.queryDiscordUser(ctx, oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	roles, err := x.queryUserGuildRoles(ctx, discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var ok bool
	for _, role := range roles {
		if role == x.Config.RoleVVGOMember {
			ok = true
			break
		}
	}
	if !ok {
		handleError(ErrNotAMember)
		return
	}

	// create the identity object
	identity := sessions.Identity{
		Kind:        sessions.IdentityDiscord,
		Roles:       []access.Role{access.RoleVVGOMember},
		DiscordUser: &sessions.DiscordUser{UserID: discordUser.ID},
	}
	loginRedirect(newCookie(ctx, x.Sessions, &identity), w, r, "/")
}

// performs the oauth request and returns the token we get from discord
func (x DiscordOAuthHandler) queryDiscordOauth(ctx context.Context, code string) (*oauthToken, error) {
	if code == "" {
		return nil, ErrInvalidOAuthCode
	}

	// build the authorization request
	form := make(url.Values)
	form.Add("client_id", x.Config.OAuthClientID)
	form.Add("client_secret", x.Config.OAuthClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", x.Config.OAuthRedirectURI)
	form.Add("scope", "identify")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://discordapp.com/api/v6/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %v", err)
	}

	// do the request
	var oauthToken oauthToken
	if _, err := doDiscordRequest(req, &oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// query discord for the token's user name
// this requires the identity scope
func (x DiscordOAuthHandler) queryDiscordUser(ctx context.Context, oauthToken *oauthToken) (*discordUser, error) {
	req, err := newTokenRequest(ctx, oauthToken, x.Config.Endpoint+"/users/@me")
	if err != nil {
		return nil, err
	}

	var discordUser discordUser
	if _, err := doDiscordRequest(req, &discordUser); err != nil {
		return nil, err
	}
	return &discordUser, nil
}

// query discord for the guild roles of the user id
// here we can use the the server's own auth token
func (x DiscordOAuthHandler) queryUserGuildRoles(ctx context.Context, userID string) ([]string, error) {
	url := fmt.Sprintf("%s/guilds/%s/members/%s", x.Config.Endpoint, x.Config.GuildID, userID)
	req, err := newBotRequest(ctx, x.Config.BotAuthToken, url)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+x.Config.BotAuthToken)

	// unmarshal the response
	var guildMember struct {
		Nick  string   `json:"nick"`
		Roles []string `json:"roles"`
	}
	if _, err := doDiscordRequest(req, &guildMember); err != nil {
		return nil, err
	}
	return guildMember.Roles, nil
}

// returns a request using an oauth token for authentication
func newTokenRequest(ctx context.Context, oauthToken *oauthToken, url string) (*http.Request, error) {
	// build the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.AccessToken))
	return req, err
}

// returns a request using a bot token for authentication
func newBotRequest(ctx context.Context, token string, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+token)
	return req, err
}

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
			"body":   buf.String(),
		}).Error("non-200 response from discord")

	default:
		logger.WithFields(logrus.Fields{
			"method": req.Method,
			"status": resp.StatusCode,
		}).Info("discord api request complete")
		err = json.NewDecoder(resp.Body).Decode(dest)
		if err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
	}
	return resp, err
}
