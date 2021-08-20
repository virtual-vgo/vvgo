package server

import (
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func ProjectsView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := sheets.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.MethodFailure(ctx, "sheets.ListProjects", err)
		helpers.InternalServerError(w)
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

func serveProject(w http.ResponseWriter, r *http.Request, project sheets.Project) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		helpers.MethodNotAllowed(w)
		return
	}

	credits, err := sheets.ListCredits(ctx)
	if err != nil {
		logger.WithError(err).Error("valuesToCredits() failed")
		helpers.InternalServerError(w)
		return
	}

	type minorTable struct {
		Name string
		Rows []sheets.Credit
	}

	type majorTable struct {
		Name   string
		Rows   []*minorTable
		rowMap map[string]*minorTable
	}

	var creditsTable struct {
		Rows   []*majorTable
		rowMap map[string]*majorTable
	}
	creditsTable.rowMap = make(map[string]*majorTable)

	for _, projectCredit := range credits.ForProject(project.Name) {
		if creditsTable.rowMap[projectCredit.MajorCategory] == nil {
			creditsTable.rowMap[projectCredit.MajorCategory] = new(majorTable)
			creditsTable.rowMap[projectCredit.MajorCategory].Name = projectCredit.MajorCategory
			creditsTable.rowMap[projectCredit.MajorCategory].rowMap = make(map[string]*minorTable)
			creditsTable.Rows = append(creditsTable.Rows, creditsTable.rowMap[projectCredit.MajorCategory])
		}
		major := creditsTable.rowMap[projectCredit.MajorCategory]

		if major.rowMap[projectCredit.MinorCategory] == nil {
			major.rowMap[projectCredit.MinorCategory] = new(minorTable)
			major.rowMap[projectCredit.MinorCategory].Name = projectCredit.MinorCategory
			major.Rows = append(major.Rows, major.rowMap[projectCredit.MinorCategory])
		}
		minor := major.rowMap[projectCredit.MinorCategory]

		minor.Rows = append(minor.Rows, projectCredit)
	}

	page := struct {
		sheets.Project
		Credits []*majorTable
	}{
		Project: project,
		Credits: creditsTable.Rows,
	}

	ParseAndExecute(ctx, w, r, &page, "project.gohtml")
}
