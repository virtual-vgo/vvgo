package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"io"
	"net/http"
	"time"
)

func Sessions(r *http.Request) api.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)
	switch r.Method {
	case http.MethodGet:
		return handleGetSession(ctx, identity)
	case http.MethodDelete:
		return handleDeleteSessions(r, ctx)
	case http.MethodPost:
		return handlePostSessions(r.Body, identity, ctx)
	default:
		return errors.NewMethodNotAllowedError()
	}
}

func handleGetSession(ctx context.Context, identity auth.Identity) api.Response {
	sessions, err := auth.ListSessions(ctx, identity)
	if err != nil {
		logger.MethodFailure(ctx, "models.ListSessions", err)
		return errors.NewInternalServerError()
	}
	return api.Response{Status: api.StatusOk, Sessions: sessions}
}

type DeleteSessionsRequest struct {
	Sessions []string `json:"sessions"`
}

func handleDeleteSessions(r *http.Request, ctx context.Context) api.Response {
	var data DeleteSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return errors.NewJsonDecodeError(err)
	}

	if len(data.Sessions) == 0 {
		return errors.NewBadRequestError("sessions must not be empty")
	}

	sessionIds := make([]string, 0, len(data.Sessions))
	for _, session := range data.Sessions {
		sessionIds = append(sessionIds, "sessions:"+session)
	}
	if err := redis.Do(ctx, redis.Cmd(nil, "DEL", sessionIds...)); err != nil {
		logger.RedisFailure(ctx, err)
		return errors.NewInternalServerError()
	}

	return api.NewOkResponse()
}

type PostSessionsRequest struct {
	Sessions []struct {
		Kind    string   `json:"kind"`
		Roles   []string `json:"roles"`
		Expires int      `json:"expires"`
	} `json:"sessions"`
}

func handlePostSessions(body io.Reader, identity auth.Identity, ctx context.Context) api.Response {
	var data PostSessionsRequest
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return errors.NewJsonDecodeError(err)
	}

	newSessionId := func(roles []string) auth.Identity {
		roleForRole := map[auth.Role]auth.Role{
			auth.RoleWriteSpreadsheet: auth.RoleVVGOExecutiveDirector,
			auth.RoleReadSpreadsheet:  auth.RoleVVGOProductionTeam,
			auth.RoleDownload:         auth.RoleVVGOVerifiedMember,
		}
		var allowed []auth.Role
		for i := range roles {
			role := auth.Role(roles[i])
			if identity.HasRole(roleForRole[role]) {
				allowed = append(allowed, role)
			}
		}
		return auth.Identity{Kind: auth.KindApiToken, Roles: allowed, DiscordID: identity.DiscordID}
	}

	var results []auth.Identity
	for i, sessionData := range data.Sessions {
		newIdentity := newSessionId(sessionData.Roles)
		if len(newIdentity.Roles) == 0 {
			return errors.NewBadRequestError(fmt.Sprintf("session %d has no usable roles", i))
		}

		var expires time.Duration
		switch {
		case sessionData.Expires == 0:
			expires = 24 * 3600 * time.Second
		case time.Duration(sessionData.Expires)*time.Second < 5*time.Second:
			expires = 5 * time.Second
		default:
			expires = time.Duration(sessionData.Expires) * time.Second
		}

		var err error
		_, err = auth.NewSession(ctx, &newIdentity, expires)
		if err != nil {
			logger.MethodFailure(ctx, "login.NewSession", err)
			return errors.NewInternalServerError()
		}
		results = append(results, newIdentity)
	}

	return api.Response{
		Status:   api.StatusOk,
		Sessions: results,
	}
}
