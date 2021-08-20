package roles

import (
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	identity := login.IdentityFromContext(r.Context())
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
