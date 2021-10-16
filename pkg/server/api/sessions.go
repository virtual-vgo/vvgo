package api

import (
	"encoding/json"
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
	switch r.Method {
	case http.MethodGet:
		identity := login.IdentityFromContext(ctx)
		sessions, err := models.ListSessions(ctx, *identity)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListSessions", err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
		json.NewEncoder(w).Encode(sessions)

	case http.MethodDelete:
		var data models.DeleteSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http_helpers.JsonDecodeFailure(ctx, w, err)
			return
		}

		if len(data.Sessions) == 0 {
			http_helpers.BadRequest(ctx, w, "sessions must not be empty")
			return
		}

		sessions := make([]string, 0, len(data.Sessions))
		for _, session := range data.Sessions {
			sessions = append(sessions, "sessions:"+session)
		}
		if err := redis.Do(ctx, redis.Cmd(nil, "DEL", sessions...)); err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.InternalServerError(ctx, w)
		}

		http_helpers.WriteAPIResponse(ctx, w, models.Response{
			Status:   models.StatusOk,
			Type:     models.ResponseTypeSessions,
			Sessions: &models.SessionsResponse{Type: models.SessionResponseTypeDeleted},
		})

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.MethodFailure(ctx, "r.ParseForm", err)
			http_helpers.BadRequest(ctx, w, "could not parse form")
			return
		}
		wantRoles := r.Form["with_roles"]
		allowedRoles := []models.Role{
			models.RoleReadConfig,
			models.RoleWriteSpreadsheet,
		}
		var roles []models.Role
		for _, want := range wantRoles {
			for _, allowed := range allowedRoles {
				if want == allowed.String() {
					roles = append(roles, allowed)
				}
			}
		}
		if len(roles) == 0 {
			http_helpers.BadRequest(ctx, w, "no roles for identity")
			return
		}

		expires := 24 * 3600 * time.Second
		identity := models.Identity{Kind: "SessionToken", Roles: roles}
		session, err := login.NewSession(ctx, &identity, expires)
		if err != nil {
			logger.MethodFailure(ctx, "login.NewSession", err)
		}
		data := make(map[string]interface{})
		data["roles"] = identity.Roles
		data["session"] = session
		data["expires"] = time.Now().Add(expires)
		w.Header().Set("HtmlSource-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.JsonEncodeFailure(ctx, err)
		}
	default:
		http_helpers.MethodNotAllowed(ctx, w)
	}
}
