package views

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Projects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	name := r.FormValue("name")
	project, ok := projects.Get(name)
	if !ok {
		ParseAndExecute(ctx, w, r, &projects, "projects_index.gohtml")
	} else {
		serveProject(w, r, project)
	}
}

func serveProject(w http.ResponseWriter, r *http.Request, project models.Project) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http_helpers.WriteErrorMethodNotAllowed(ctx, w)
		return
	}

	credits, err := models.ListCredits(ctx)
	if err != nil {
		logger.ListCreditsFailure(ctx, err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	page := struct {
		models.Project
		Credits models.CreditsTable
	}{
		Project: project,
		Credits: models.BuildCreditsTable(credits, project),
	}

	ParseAndExecute(ctx, w, r, &page, "project.gohtml")
}
