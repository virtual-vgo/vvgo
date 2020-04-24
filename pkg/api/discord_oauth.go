package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type DiscordOAuthHandlerConfig struct {
	Endpoint          string `default:"https://discordapp.com/api/v6"`
	AuthToken         string `split_words:"true"`
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
var ErrOAuthRequestFailed = errors.New("oauth request failed")

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

	buf, _ := httputil.DumpRequest(r, true)
	fmt.Println(string(buf))

	var identity sessions.Identity
	if err := func() error {
		code := r.FormValue("code")
		req, err := x.buildOAuthRequest(ctx, code)
		if err != nil {
			return err
		}

		oauthToken, err := x.doOAuthRequest(req)
		if err != nil {
			return err
		}

		discordUser, err := x.queryDiscordUser(ctx, oauthToken)
		if err != nil {
			return err
		}

		roles, err := x.queryUserGuildRoles(ctx, discordUser.ID)
		if err != nil {
			return err
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
			return ErrNotAMember
		}

		// create the identity object
		identity = sessions.Identity{
			Kind:        sessions.IdentityDiscord,
			Roles:       []access.Role{access.RoleVVGOMember},
			DiscordUser: &sessions.DiscordUser{UserID: discordUser.ID},
		}
		return nil
	}(); err != nil {
		logger.WithError(err).Error("httpClient.Do() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// create a session and cookie
	session := x.Sessions.NewSession(time.Now().Add(7 * 24 * 3600 * time.Second))
	cookie := x.Sessions.NewCookie(session)
	if err := x.Sessions.StoreIdentity(ctx, session.ID, &identity); err != nil {
		logger.WithError(err).Error("x.Sessions.Save() failed")
		internalServerError(w)
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

var ErrDiscordRequestFailed = errors.New("discord request failed")

// builds the oauth request for identity access
func (x DiscordOAuthHandler) buildOAuthRequest(ctx context.Context, code string) (*http.Request, error) {
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
	return req, err
}

// performs the oauth request and returns the token
func (x DiscordOAuthHandler) doOAuthRequest(req *http.Request) (*oauthToken, error) {
	buf, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(buf))
	// do the request
	resp, err := tracing.DoHttpRequest(req)
	buf, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(buf))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.WithField("status", resp.StatusCode).Error("non-200 response from discord")
		return nil, ErrOAuthRequestFailed
	}

	// unmarshal the response
	var oauthToken oauthToken
	if err := json.NewDecoder(resp.Body).Decode(&oauthToken); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil, err
	}
	return &oauthToken, nil
}

// query discord for the token's user name
func (x DiscordOAuthHandler) queryDiscordUser(ctx context.Context, oauthToken *oauthToken) (*discordUser, error) {
	// build the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, x.Config.Endpoint+"/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", oauthToken.TokenType+" "+oauthToken.AccessToken)

	// do the request
	buf, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(buf))
	resp, err := tracing.DoHttpRequest(req)
	buf, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(buf))
	if resp.StatusCode != http.StatusOK {
		logger.WithField("status", resp.StatusCode).Error("non-200 response from discord")
		return nil, ErrDiscordRequestFailed
	}

	// unmarshal the payload
	var discordUser discordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil, err
	}
	return &discordUser, nil
}

// query discord for the guild roles of the user id
func (x DiscordOAuthHandler) queryUserGuildRoles(ctx context.Context, userID string) ([]string, error) {
	// verify this user is in our discord server
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/guilds/%s/members/%s", x.Config.Endpoint, x.Config.GuildID, userID), nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+x.Config.AuthToken)

	buf, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(buf))
	resp, err := tracing.DoHttpRequest(req)
	buf, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(buf))
	if err != nil {
		logger.WithError(err).Error("tracing.DoHttpRequest() failed")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrDiscordRequestFailed
	}

	// unmarshal the response
	var guildMember struct {
		Nick  string   `json:"nick"`
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&guildMember); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil, err

	}
	return guildMember.Roles, nil
}
