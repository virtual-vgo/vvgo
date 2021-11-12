package credits

import (
	"fmt"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/api/website_data"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

// Table is the credits for a project.
type Table []TopicRow

// TopicRow represents either all performers or all crew.
type TopicRow struct {
	Name   string     `json:"Name"`
	Rows   []*TeamRow `json:"Rows"`
	rowMap map[string]*TeamRow
}

// TeamRow represents either a production team or instrument section.
type TeamRow struct {
	Name string   `json:"Name"`
	Rows []Credit `json:"Rows"`
}

type GetTableUrlParams struct {
	ProjectName string
}

func ServeTable(r *http.Request) http2.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)

	var data GetTableUrlParams
	data.ProjectName = r.URL.Query().Get("projectName")
	if data.ProjectName == "" {
		return errors.NewBadRequestError("projectName is required")
	}

	projects, err := website_data.ListProjects(ctx, identity)
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		return errors.NewInternalServerError()
	}
	wantProject, ok := website_data.GetProject(projects, data.ProjectName)
	if !ok {
		return errors.NewNotFoundError(fmt.Sprintf("project %s not found", data.ProjectName))
	}

	credits, err := ListCredits(ctx)
	if err != nil {
		logger.ListCreditsFailure(ctx, err)
		return errors.NewInternalServerError()
	}

	return http2.Response{Status: http2.StatusOk, CreditsTable: BuildCreditsTable(credits, wantProject)}
}

func BuildCreditsTable(credits Credits, project website_data.Project) Table {
	var rows []*TopicRow
	rowMap := make(map[string]*TopicRow)
	for _, projectCredit := range credits.ForProject(project.Name) {
		if rowMap[projectCredit.MajorCategory] == nil {
			rowMap[projectCredit.MajorCategory] = new(TopicRow)
			rowMap[projectCredit.MajorCategory].Name = projectCredit.MajorCategory
			rowMap[projectCredit.MajorCategory].rowMap = make(map[string]*TeamRow)
			rows = append(rows, rowMap[projectCredit.MajorCategory])
		}
		major := rowMap[projectCredit.MajorCategory]
		if major.rowMap[projectCredit.MinorCategory] == nil {
			major.rowMap[projectCredit.MinorCategory] = new(TeamRow)
			major.rowMap[projectCredit.MinorCategory].Name = projectCredit.MinorCategory
			major.Rows = append(major.Rows, major.rowMap[projectCredit.MinorCategory])
		}
		minor := major.rowMap[projectCredit.MinorCategory]
		minor.Rows = append(minor.Rows, projectCredit)
	}

	table := make(Table, len(rows))
	for i := range rows {
		table[i] = *rows[i]
	}
	return table
}
