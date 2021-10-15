package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	login2 "github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"time"
)

func Session(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		logger.MethodFailure(ctx, "r.ParseForm", err)
		helpers.BadRequest(w, "could not parse form")
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
		helpers.BadRequest(w, "no roles for identity")
		return
	}

	expires := 24 * 3600 * time.Second
	identity := models.Identity{Kind: "SessionToken", Roles: roles}
	session, err := login2.NewSession(ctx, &identity, expires)
	if err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
	}
	data := make(map[string]interface{})
	data["roles"] = identity.Roles
	data["session"] = session
	data["expires"] = time.Now().Add(expires)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}
