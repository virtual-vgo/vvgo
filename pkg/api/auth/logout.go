package auth

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"net/http"
)

func Logout(r *http.Request) api.Response {
	ctx := r.Context()
	identity := IdentityFromContext(ctx)
	if err := DeleteSession(ctx, identity.Key); err != nil {
		return response.NewInternalServerError()
	}
	return api.NewOkResponse()
}
