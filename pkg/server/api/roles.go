package api

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Roles(w http.ResponseWriter, r *http.Request) {
	identity := login.IdentityFromContext(r.Context())
	var wantRoles []models.Role
	http_helpers.JsonDecode(r.Body, &wantRoles)
	var roles []models.Role
	if len(wantRoles) != 0 {
		roles = identity.AssumeRoles(wantRoles...).Roles
	} else {
		roles = identity.Roles
	}
	http_helpers.JsonEncode(w, roles)
}
