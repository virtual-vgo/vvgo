package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets/credit"
	"github.com/virtual-vgo/vvgo/pkg/sheets/project"
	"net/http"
)

type ProjectsView struct{ Template }

func (x ProjectsView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/projects/" {
		x.serveIndex(w, r)
	} else {
		x.serveProject(w, r, r.URL.Path[len("/projects/"):])
	}
}

func (x ProjectsView) serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := project.List(ctx, IdentityFromContext(ctx), x.SpreadsheetID)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	x.Template.ParseAndExecute(ctx, w, r, &projects, "projects/index.gohtml")
}

func (x ProjectsView) serveProject(w http.ResponseWriter, r *http.Request, name string) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := project.List(ctx, IdentityFromContext(ctx), x.SpreadsheetID)
	if err != nil {
		logger.WithError(err).Error("valuesToProjects() failed")
		internalServerError(w)
		return
	}

	wantProject, ok := projects.WithName(name)
	if !ok {
		http.NotFound(w, r)
		return
	}

	credits, err := credit.List(ctx, x.SpreadsheetID)
	if err != nil {
		logger.WithError(err).Error("valuesToCredits() failed")
		internalServerError(w)
		return
	}

	type minorTable struct {
		Name string
		Rows []credit.Credit
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

	for _, projectCredit := range credits.ForProject(wantProject.Name) {
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
		project.Project
		Credits []*majorTable
	}{
		Project: wantProject,
		Credits: creditsTable.Rows,
	}

	x.Template.ParseAndExecute(ctx, w, r, &page, "projects/project.gohtml")
}
