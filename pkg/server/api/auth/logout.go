package auth

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Logout(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)
	if err := login.DeleteSession(ctx, identity.Key); err != nil {
		return http_helpers.NewInternalServerError()
	}
	return http_helpers.NewOkResponse()
}
