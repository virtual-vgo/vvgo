package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func ProjectsApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	projects, err := sheets.ListProjects(ctx, IdentityFromContext(ctx))
	if err != nil {
		logger.WithError(err).Error("valuesToProjects() failed")
		internalServerError(w)
		return
	}
	switch {
	case r.FormValue("latest") == "true":
		project := projects.WithField("Video Released", true).Sort().Last()
		handleError(json.NewEncoder(w).Encode(sheets.Projects{project})).
			logError("json.Encode() failed")
	default:
		handleError(json.NewEncoder(w).Encode(projects)).
			logError("json.Encode() failed")
	}
}

func ProjectsView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := sheets.ListProjects(ctx, IdentityFromContext(ctx))
	handleError(err).ifError(func(err error) {
		logger.WithError(err).Error("sheets.ListProjects() failed")
		internalServerError(w)
	}).ifSuccess(func() {
		name := r.FormValue("name")
		project, ok := projects.Get(name)
		if !ok {
			ParseAndExecute(ctx, w, r, &projects, "projects_index.gohtml")
		} else {
			serveProject(w, r, project)
		}
	})
}

func serveProject(w http.ResponseWriter, r *http.Request, project sheets.Project) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	credits, err := sheets.ListCredits(ctx)
	if err != nil {
		logger.WithError(err).Error("valuesToCredits() failed")
		internalServerError(w)
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
