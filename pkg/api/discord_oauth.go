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
	"io"
	"net/http"
	"net/url"
	"strings"
)

type DiscordOAuthHandlerConfig struct {
	GuildID           string `split_words:"true"`
	RoleVVGOMember    string `envconfig:"role_vvgo_member"`
}

type DiscordOAuthHandler struct {
	Config   DiscordOAuthHandlerConfig
	Sessions *sessions.Store
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

