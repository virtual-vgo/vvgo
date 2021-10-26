package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"io"
	"net/http"
	"time"
)

func Sessions(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)
	switch r.Method {
	case http.MethodGet:
		return handleGetSession(ctx, identity)
	case http.MethodDelete:
		return handleDeleteSessions(r, ctx)
	case http.MethodPost:
		return handlePostSessions(r.Body, identity, ctx)
	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

func handleGetSession(ctx context.Context, identity models.Identity) models.ApiResponse {
	sessions, err := models.ListSessions(ctx, identity)
	if err != nil {
		logger.MethodFailure(ctx, "models.ListSessions", err)
		return http_helpers.NewInternalServerError()
	}
	return models.ApiResponse{Status: models.StatusOk, Sessions: sessions}
}

type DeleteSessionsRequest struct {
	Sessions []string `json:"sessions"`
}

func handleDeleteSessions(r *http.Request, ctx context.Context) models.ApiResponse {
	var data DeleteSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	if len(data.Sessions) == 0 {
		return http_helpers.NewBadRequestError("sessions must not be empty")
	}

	sessionIds := make([]string, 0, len(data.Sessions))
	for _, session := range data.Sessions {
		sessionIds = append(sessionIds, "sessions:"+session)
	}
	if err := redis.Do(ctx, redis.Cmd(nil, "DEL", sessionIds...)); err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	return http_helpers.NewOkResponse()
}

type PostSessionsRequest struct {
	Sessions []struct {
		Kind    string   `json:"kind"`
		Roles   []string `json:"roles"`
		Expires int      `json:"expires"`
	} `json:"sessions"`
}

func handlePostSessions(body io.Reader, identity models.Identity, ctx context.Context) models.ApiResponse {
	var data PostSessionsRequest
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	newSessionId := func(roles []string) models.Identity {
		allowedRoles := map[models.Role]bool{
			models.RoleReadConfig:       true,
			models.RoleWriteSpreadsheet: true,
		}
		var allowed []models.Role
		for _, role := range roles {
			if allowedRoles[models.Role(role)] {
				allowed = append(allowed, models.Role(role))
			}
		}
		return models.Identity{Kind: models.KindApiToken, Roles: allowed, DiscordID: identity.DiscordID}
	}

	var results []models.Identity
	for i, sessionData := range data.Sessions {
		newIdentity := newSessionId(sessionData.Roles)
		if len(newIdentity.Roles) == 0 {
			return http_helpers.NewBadRequestError(fmt.Sprintf("session %d has no usable roles", i))
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
		_, err = login.NewSession(ctx, &newIdentity, expires)
		if err != nil {
			logger.MethodFailure(ctx, "login.NewSession", err)
			return http_helpers.NewInternalServerError()
		}
		results = append(results, newIdentity)
	}

	return models.ApiResponse{
		Status:   models.StatusOk,
		Sessions: results,
	}
}
