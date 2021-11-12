package auth

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"net/http"
)

func Logout(r *http.Request) api.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)
	if err := auth.DeleteSession(ctx, identity.Key); err != nil {
		return errors.NewInternalServerError()
	}
	return api.NewOkResponse()
}
