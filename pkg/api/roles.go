package api

import (
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
)

func RolesApi(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	var wantRoles []login.Role
	helpers.JsonDecode(r.Body, &wantRoles)
	var roles []login.Role
	if len(wantRoles) != 0 {
		roles = identity.AssumeRoles(wantRoles...).Roles
	} else {
		roles = identity.Roles
	}
	helpers.JsonEncode(w, roles)
}
