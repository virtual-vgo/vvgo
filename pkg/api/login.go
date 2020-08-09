package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/facebook"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

const LoginCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks

type LoginView struct {
	Sessions *login.Store
}

func (x LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	identity := identityFromContext(ctx)
	if identity.IsAnonymous() == false {
		http.Redirect(w, r, "/login/success", http.StatusFound)
		return
	}

	opts := NewNavBarOpts(ctx)
	opts.LoginActive = true
	page := struct {
		NavBar NavBarOpts
	}{
		NavBar: opts,
	}

	var buf bytes.Buffer
	if ok := parseAndExecute(&buf, &page, filepath.Join(PublicFiles, "login.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buf.WriteTo(w)
}

type LoginSuccessView struct{}

func (LoginSuccessView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opts := NewNavBarOpts(ctx)
	page := struct {
		NavBar NavBarOpts
	}{
		NavBar: opts,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "login_success.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func loginSuccess(w http.ResponseWriter, r *http.Request, ctx context.Context, sessions *login.Store, identity *login.Identity) {
	cookie, err := sessions.NewCookie(ctx, identity, LoginCookieDuration)
	if err != nil {
		logger.WithError(err).Error("store.NewCookie() failed")
		internalServerError(w)
		return
	}

	http.SetCookie(w, cookie)
	logger.WithFields(logrus.Fields{
		"identity": identity.Kind,
		"roles":    identity.Roles,
	}).Info("authorization succeeded")

	LoginSuccessView{}.ServeHTTP(w, r)
}

// PasswordLoginHandler authenticates requests using form values user and pass and a static map of valid combinations.
// If the user pass combo exists in the map, then a login cookie with the mapped roles is create and sent in the response.
type PasswordLoginHandler struct {
	Sessions *login.Store

	// Logins is a map of login user and pass to a slice of roles for that login.
	Logins map[[2]string][]login.Role
}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var identity login.Identity
	if err := x.Sessions.ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/parts", http.StatusFound)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	gotRoles, ok := x.Logins[[2]string{user, pass}]
	if !ok {
		logger.WithFields(logrus.Fields{
			"user": user,
		}).Error("password authentication failed")
		unauthorized(w)
		return
	}

	loginSuccess(w, r, ctx, x.Sessions, &login.Identity{
		Kind:  login.KindPassword,
		Roles: gotRoles,
	})
}

const DiscordOAuthPreCookie = "vvgo-discord-oauth-pre"

// DiscordLoginHandler
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct {
	GuildID        discord.GuildID
	RoleVVGOMember string
	Sessions       *login.Store
	Namespace      string
	RedirectURL    string
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") == "" {
		x.redirect(w, r)
	} else {
		x.authorize(w, r)
	}
}

func (x DiscordLoginHandler) redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// read a random state number
	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.WithError(err).Error("rand.Read() failed")
		internalServerError(w)
		return
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	value := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	// store the number in redis
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", x.Namespace+":discord_oauth_pre:"+state, "300", value)); err != nil {
		logger.WithError(err).Error("redis.Do() failed")
		internalServerError(w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    DiscordOAuthPreCookie,
		Value:   value,
		Expires: time.Now().Add(300 * time.Second),
	})
	redirectURL, err := url.Parse(x.RedirectURL)
	if err != nil {
		logger.WithError(err).Error("url.Parse() failed")
		internalServerError(w)
		return
	}
	query := redirectURL.Query()
	query.Set("state", state)
	redirectURL.RawQuery = query.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

func (x DiscordLoginHandler) authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("discord authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	// read the state param
	state := r.FormValue("state")
	if state == "" {
		handleError(errors.New("no state param"))
		return
	}

	// check if it exists in redis
	var value string
	if ok := handleError(redis.Do(ctx, redis.Cmd(&value, "GET", x.Namespace+":discord_oauth_pre:"+state))); !ok {
		return
	}

	// check against the cookie value
	preCookie, err := r.Cookie(DiscordOAuthPreCookie)
	if ok := handleError(err); !ok {
		return
	}
	if preCookie.Value != value {
		handleError(errors.New("invalid state"))
		return
	}

	// get an oauth token from discord
	code := r.FormValue("code")
	oauthToken, err := discord.QueryOAuth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	// get the user id
	discordUser, err := discord.QueryIdentity(ctx, oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	// check if this user is in our guild
	guildMember, err := discord.QueryGuildMember(ctx, x.GuildID, discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var ok bool
	for _, role := range guildMember.Roles {
		if role == x.RoleVVGOMember {
			ok = true
			break
		}
	}
	if !ok {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, ctx, x.Sessions, &login.Identity{
		Kind:  login.KindDiscord,
		Roles: []login.Role{login.RoleVVGOMember},
	})
}

const FacebookOAuthPreCookie = "vvgo-facebook-oauth-pre"

// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow/

type FacebookLoginHandler struct {
	VVGOGroupID string
	Namespace   string
	RedirectURL string
	Sessions    *login.Store
}

func (x FacebookLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") == "" {
		x.redirect(w, r)
	} else {
		x.authorize(w, r)
	}
}

func (x FacebookLoginHandler) redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// read a random state number
	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.WithError(err).Error("rand.Read() failed")
		internalServerError(w)
		return
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	value := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	// store the number in redis
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", x.Namespace+":facebook_oauth_pre:"+state, "300", value)); err != nil {
		logger.WithError(err).Error("redis.Do() failed")
		internalServerError(w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    FacebookOAuthPreCookie,
		Value:   value,
		Expires: time.Now().Add(300 * time.Second),
	})
	http.Redirect(w, r, facebook.LoginURL(state), http.StatusFound)
}

func (x FacebookLoginHandler) authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("facebook authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	// read the state param
	state := r.FormValue("state")
	if state == "" {
		handleError(errors.New("no state param"))
		return
	}

	// check if it exists in redis
	var value string
	if ok := handleError(redis.Do(ctx, redis.Cmd(&value, "GET", x.Namespace+":facebook_oauth_pre:"+state))); !ok {
		return
	}

	// check against the cookie value
	preCookie, err := r.Cookie(FacebookOAuthPreCookie)
	if ok := handleError(err); !ok {
		return
	}
	if preCookie.Value != value {
		handleError(errors.New("invalid state"))
		return
	}

	// get an oauth token from facebook
	code := r.FormValue("code")
	oauthToken, err := facebook.QueryOAuth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	isMember, err := facebook.UserHasGroup(ctx, oauthToken.AccessToken, "me", x.VVGOGroupID)
	if ok := handleError(err); !ok {
		return
	}
	if !isMember {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, ctx, x.Sessions, &login.Identity{
		Kind:  login.KindFacebook,
		Roles: []login.Role{login.RoleVVGOMember},
	})
}

// LogoutHandler deletes the login session from the incoming request, if it exists.
type LogoutHandler struct {
	Sessions *login.Store
}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := x.Sessions.DeleteSessionFromRequest(ctx, r); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteSessionFromRequest failed")
		internalServerError(w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
