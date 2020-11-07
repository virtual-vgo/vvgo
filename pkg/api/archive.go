package api

import (
	"bytes"
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
)

type ArchiveView struct {
	SpreadsheetID string
	*Database
}

func (x ArchiveView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/archive/" {
		x.serveIndex(w, r)
	} else {
		x.serveProject(w, r, r.URL.Path[len("/archive/"):])
	}
}

func (x ArchiveView) serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projectValues, err := readSheet(ctx, x.SpreadsheetID, ProjectsRange)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	projects := listProjects(projectValues)

	projects = x.filterFromQuery(r, projects)
	x.renderIndexView(w, ctx, projects)
}

func (x ArchiveView) filterFromQuery(r *http.Request, projects []Project) []Project {
	identity := identityFromContext(r.Context())
	want := len(projects)
	for i := 0; i < want; i++ {
		if projects[i].Released == true || identity.HasRole(login.RoleVVGOTeams) || identity.HasRole(login.RoleVVGOLeader) {
			continue
		}
		projects[i], projects[want-1] = projects[want-1], projects[i]
		i--
		want--
	}
	projects = projects[:want]
	return projects
}

func (x ArchiveView) renderIndexView(w http.ResponseWriter, ctx context.Context, projects []Project) {
	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &projects, "archive/index.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func (x ArchiveView) serveProject(w http.ResponseWriter, r *http.Request, name string) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	values, err := readSheet(ctx, x.SpreadsheetID, ProjectsRange)
	if err != nil {
		logger.WithError(err).Error("listProjects() failed")
		internalServerError(w)
		return
	}
	projects := listProjects(values)
	var exists bool
	var wantProject Project
	for _, project := range projects {
		if project.Name == name {
			exists = true
			wantProject = project
			break
		}
	}
	if !exists {
		http.NotFound(w, r)
		return
	}
	renderProjectView(w, ctx, wantProject, x.SpreadsheetID)
}

func renderProjectView(w http.ResponseWriter, ctx context.Context, project Project, spreadsheetID string) {
	values, err := readSheet(ctx, spreadsheetID, CreditsRange)
	if err != nil {
		logger.WithError(err).Error("listCredits() failed")
		internalServerError(w)
		return
	}

	credits := listCredits(values)

	type minorTable struct {
		Name string
		Rows []*Credit
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

	for i := range credits {
		if credits[i].Project != project.Name {
			continue
		}

		if creditsTable.rowMap[credits[i].MajorCategory] == nil {
			creditsTable.rowMap[credits[i].MajorCategory] = new(majorTable)
			creditsTable.rowMap[credits[i].MajorCategory].Name = credits[i].MajorCategory
			creditsTable.rowMap[credits[i].MajorCategory].rowMap = make(map[string]*minorTable)
			creditsTable.Rows = append(creditsTable.Rows, creditsTable.rowMap[credits[i].MajorCategory])
		}
		major := creditsTable.rowMap[credits[i].MajorCategory]

		if major.rowMap[credits[i].MinorCategory] == nil {
			major.rowMap[credits[i].MinorCategory] = new(minorTable)
			major.rowMap[credits[i].MinorCategory].Name = credits[i].MinorCategory
			major.Rows = append(major.Rows, major.rowMap[credits[i].MinorCategory])
		}
		minor := major.rowMap[credits[i].MinorCategory]

		minor.Rows = append(minor.Rows, &credits[i])
	}

	page := struct {
		Project
		Credits []*majorTable
	}{
		Project: project,
		Credits: creditsTable.Rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &page, "archive/project.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
