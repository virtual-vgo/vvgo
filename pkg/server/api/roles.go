package api

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	login2 "github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Roles(w http.ResponseWriter, r *http.Request) {
	identity := login2.IdentityFromContext(r.Context())
	var wantRoles []models.Role
	helpers.JsonDecode(r.Body, &wantRoles)
	var roles []models.Role
	if len(wantRoles) != 0 {
		roles = identity.AssumeRoles(wantRoles...).Roles
	} else {
		roles = identity.Roles
	}
	helpers.JsonEncode(w, roles)
}
