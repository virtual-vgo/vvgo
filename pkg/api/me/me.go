package me

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"net/http"
)

func Me(r *http.Request) api.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)
	return api.Response{
		Status:   api.StatusOk,
		Identity: &identity,
	}
}
