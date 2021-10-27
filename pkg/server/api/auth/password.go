package auth

import (
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type PostPasswordRequest struct {
	User string
	Pass string
}

func Password(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		return http_helpers.NewMethodNotAllowedError()
	}

	passwords := make(map[string]string)
	passwords["vvgo-member"] = config.Config.VVGO.MemberPasswordHash

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
		return http_helpers.NewUnauthorizedError()
	}

	identity := models.Identity{
		Kind:  models.KindPassword,
		Roles: []models.Role{models.RoleVVGOVerifiedMember},
	}
	if _, err := login.NewSession(ctx, &identity, SessionDuration); err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
		return http_helpers.NewInternalServerError()
	}

	return models.ApiResponse{Status: models.StatusOk, Identity: &identity}
}
