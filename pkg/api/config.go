package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"io"
	"net/http"
	"os"
	"time"
)

var SessionApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		logger.MethodFailure(ctx, "r.ParseForm", err)
		badRequest(w, "could not parse form")
		return
	}
	wantRoles := r.Form["with_roles"]
	allowedRoles := []login.Role{login.RoleReadConfig}
	var roles []login.Role
	for _, want := range wantRoles {
		for _, allowed := range allowedRoles {
			if want == allowed.String() {
				roles = append(roles, allowed)
			}
		}
	}
	if len(roles) == 0 {
		badRequest(w, "no roles for identity")
		return
	}

	expires := 24 * 3600 * time.Second
	identity := login.Identity{Kind: "SessionToken", Roles: roles}
	session, err := login.NewSession(ctx, &identity, expires)
	if err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
	}
	data := make(map[string]interface{})
	data["roles"] = identity.Roles
	data["session"] = session
	data["expires"] = time.Now().Add(expires)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
})

var ConfigApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if parse_config.Session != "" {
		methodNotAllowed(w)
		return
	}

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	ctx := r.Context()
	config := make(map[string]interface{})
	file, err := os.Open(parse_config.FileName)
	if err != nil {
		logger.MethodFailure(ctx, "os.Open", err)
		internalServerError(w)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		logger.MethodFailure(ctx, "io.Copy", err)
		internalServerError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
})
