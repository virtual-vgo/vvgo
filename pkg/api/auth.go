package api

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/users"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const HeaderVirtualVGOApiToken = "Virtual-VGO-Api-Secret"

type PassThrough struct{}

func (x PassThrough) Authenticate(handler http.Handler) http.Handler {
	return handler
}

// Authenticates http requests using basic auth.
// Identity name is the map key, and password is the value.
// If the map is empty or nil, requests are always authenticated.
type BasicAuth map[string]string

func (auth BasicAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ok := func() bool {
			_, span := tracing.StartSpan(ctx, "basic_auth")
			defer span.Send()
			user, pass, ok := r.BasicAuth()
			if !ok || user == "" || pass == "" {
				return false
			}
			if auth[user] == pass {
				return true
			} else {
				logger.WithFields(logrus.Fields{
					"user": user,
				}).Error("user authentication failed")
				return false
			}
		}(); ok {
			tracing.WrapHandler(handler).ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			unauthorized(w)
		}
	})
}

type TokenAuth []string

func (tokens TokenAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ok := func() bool {
			_, span := tracing.StartSpan(ctx, "token_auth")
			defer span.Send()
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			for _, token := range tokens {
				if auth == "Bearer "+token {
					return true
				}
			}
			return false
		}(); ok {
			tracing.WrapHandler(handler).ServeHTTP(w, r)
		} else {
			logger.Error("token authentication failed")
			unauthorized(w)
		}
	})
}

type DiscordOAuthHandlerConfig struct {
	// Discord gives us this token when it redirects a user to our server
	OAuthDiscordSecret string
	OAuthClientID      string
	OAuthClientSecret  string
	OAuthRedirectURI   string
	GuildID            string
}

type DiscordOAuthHandler struct {
	DiscordOAuthHandlerConfig
	Sessions sessions.Store
}

func (x DiscordOAuthHandler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "discord_oauth_handler")
	defer span.Send()

	// validate the token from discord
	// this is our way to confirm the client is genuinely discord
	if x.OAuthDiscordSecret != r.FormValue("vvgo_discord_token") {
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// read the secret code
	code := r.FormValue("code")
	if code == "" {
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// build the authorization request
	form := make(url.Values)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", x.OAuthRedirectURI)
	form.Add("scope", "identify")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://discordapp.com/api/v6/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		logger.WithError(err).Error("http.NewRequest() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}
	req.SetBasicAuth(x.OAuthClientID, x.OAuthClientSecret)

	// do the request
	resp, err := tracing.DoHttpRequest(req)
	if err != nil {
		logger.WithError(err).Error("httpClient.Do() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("authorization failed")
		unauthorized(w)
	}

	// unmarshal the response
	var auth struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&auth); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// query discord for the user name
	req, err = http.NewRequestWithContext(ctx, http.MethodGet,
		DiscordEndpoint+"/users/@me", nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}
	req.Header.Add("Authorization",
		fmt.Sprintf("%s %s", auth.TokenType, auth.AccessToken))
	resp, err = tracing.DoHttpRequest(req)

	if resp.StatusCode != http.StatusOK {
		logger.Error("authorization failed")
		unauthorized(w)
	}

	// unmarshal the response
	var discordUser struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&discordUser); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// verify this user is in our discord server
	req, err = http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/guilds/%s/members/%s", DiscordEndpoint, DiscordGuildID, discordUser.ID), nil)
	if err != nil {
		logger.WithError(err).Error("http.NewRequestWithContext() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", auth.TokenType, auth.AccessToken))
	resp, err = tracing.DoHttpRequest(req)
	if err != nil {
		logger.WithError(err).Error("tracing.DoHttpRequest() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("authorization failed")
		unauthorized(w)
	}

	// unmarshal the response
	var guildMember struct {
		Nick  string   `json:"nick"`
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&guildMember); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		tracing.AddError(ctx, err)
		logger.Error("authorization failed")
		unauthorized(w)
		return
	}

	// check that they have the member role
	var ok bool
	for _, role := range guildMember.Roles {
		if role == DiscordRoleVVGOMembers {
			ok = true
			break
		}
	}
	if !ok {
		logger.Error("not a member")
		unauthorized(w)
		return
	}

	// create a user object
	user := sessions.Identity{
		Kind:        sessions.IdentityDiscord,
		DiscordUser: sessions.DiscordUser{UserID: discordUser.ID},
	}

	// create a session and cookie
	session, cookie := x.Sessions.NewSessionCookie(time.Now().Add(7*24*3600*time.Second))
	if err := x.Sessions.Save(session.ID, &user); err != nil {
		logger.WithError(err).Error("x.Sessions.Save() failed")
		internalServerError(w)
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

const DiscordGuildID = "690626216637497425"
const DiscordRoleVVGOMembers = "i dunno"
const DiscordEndpoint = "https://discordapp.com"

