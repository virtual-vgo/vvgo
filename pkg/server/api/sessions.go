package api

import (
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	login "github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"time"
)

func Sessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)
	switch r.Method {
	case http.MethodGet:
		identity := login.IdentityFromContext(ctx)
		sessions, err := models.ListSessions(ctx, identity)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListSessions", err)
			http_helpers.WriteInternalServerError(ctx, w)
			return
		}
		http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
			Status:   models.StatusOk,
			Sessions: sessions,
		})

	case http.MethodDelete:
		var data models.DeleteSessionsRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http_helpers.WriteErrorJsonDecodeFailure(ctx, w, err)
			return
		}

		if len(data.Sessions) == 0 {
			http_helpers.WriteErrorBadRequest(ctx, w, "sessions must not be empty")
			return
		}

		sessionIds := make([]string, 0, len(data.Sessions))
		for _, session := range data.Sessions {
			sessionIds = append(sessionIds, "sessions:"+session)
		}
		if err := redis.Do(ctx, redis.Cmd(nil, "DEL", sessionIds...)); err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.WriteInternalServerError(ctx, w)
		}

		http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
			Status: models.StatusOk,
		})

	case http.MethodPost:
		var data models.CreateSessionsRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http_helpers.WriteErrorJsonDecodeFailure(ctx, w, err)
			return
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
				http_helpers.WriteErrorBadRequest(ctx, w, fmt.Sprintf("session %d has no usable roles", i))
				return
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
				http_helpers.WriteInternalServerError(ctx, w)
				return
			}
			results = append(results, newIdentity)
		}

		http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
			Status:   models.StatusOk,
			Sessions: results,
		})

	default:
		http_helpers.WriteErrorMethodNotAllowed(ctx, w)
	}
}
