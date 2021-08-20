package projects

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var logger = log.New()

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	projects, err := sheets.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.WithError(err).Error("valuesToProjects() failed")
		helpers.InternalServerError(w)
		return
	}

	if r.FormValue("latest") == "true" {
		project := projects.WithField("Video Released", true).Sort().Last()
		projects = sheets.Projects{project}
	}

	if err := json.NewEncoder(w).Encode(projects); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}
