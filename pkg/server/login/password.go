package login

import (
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Password authenticates requests using form values user and pass and a static map of valid combinations.
// If the user pass combo exists in the map, then a login cookie with the mapped roles is sent in the response.
func Password(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		helpers.MethodNotAllowed(w)
		return
	}

	passwords := make(map[string]string)
	passwords["vvgo-member"] = config.Config.VVGO.MemberPasswordHash

	var identity models.Identity
	if err := ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/parts", http.StatusFound)
		return
	}

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
		helpers.Unauthorized(w)
		return
	}

	loginSuccess(w, r.WithContext(ctx), &models.Identity{
		Kind:  models.KindPassword,
		Roles: []models.Role{models.RoleVVGOMember},
	})
}
