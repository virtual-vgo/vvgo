package auth

import (
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type PostPasswordRequest struct {
	User string
	Pass string
}

func Password(r *http.Request) api.Response {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		return response.NewMethodNotAllowedError()
	}

	passwords := make(map[string]string)
	passwords["vvgo-member"] = config.Env.VVGO.MemberPasswordHash

	user := r.FormValue("user")
	pass := r.FormValue("pass")
	var err error
	switch {
	case user == "":
		err = errors.New("user is required")
	case pass == "":
		err = errors.New("password is required")
	case passwords[user] == "":
		err = errors.New("unknown user")
	default:
		err = bcrypt.CompareHashAndPassword([]byte(passwords[user]), []byte(pass))
	}

	if err != nil {
		logger.WithError(err).WithField("user", user).Error("password authentication failed")
		return response.NewUnauthorizedError()
	}

	identity := Identity{
		Kind:  KindPassword,
		Roles: []Role{RoleVVGOVerifiedMember},
	}
	if _, err := NewSession(ctx, &identity, SessionDuration); err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
		return response.NewInternalServerError()
	}

	return api.Response{Status: api.StatusOk, Identity: &identity}
}
