package website_data

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"sort"
)

const SheetProjects = "Projects"

type Project struct {
	Name                    string
	Title                   string
	Season                  string
	Hidden                  bool
	PartsReleased           bool
	PartsArchived           bool
	VideoReleased           bool
	Sources                 string
	Composers               string
	Arrangers               string
	Editors                 string
	Transcribers            string
	Preparers               string
	ClixBy                  string
	Reviewers               string
	Lyricists               string
	AdditionalContent       string
	ReferenceTrack          string
	ChoirPronunciationGuide string
	BannerLink              string
	YoutubeLink             string
	YoutubeEmbed            string
	SubmissionDeadline      string
	SubmissionLink          string
	BandcampAlbum           string
}

func ServeProjects(r *http.Request) api.Response {
	ctx := r.Context()
	projects, err := ListProjects(ctx, auth.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		return response.NewInternalServerError()
	}

	if projects == nil {
		projects = []Project{}
	}

	sort.Slice(projects, func(i, j int) bool { return projects[j].Name < projects[i].Name })
	return api.Response{Status: api.StatusOk, Projects: projects}
}

func ListProjects(ctx context.Context, identity auth.Identity) ([]Project, error) {
	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetProjects)
	if err != nil {
		return nil, err
	}

	if len(values) < 1 {
		return nil, nil
	}
	var projects []Project
	UnmarshalSheet(values, &projects)

	valid := make([]Project, 0, len(projects))
	for _, project := range projects {
		switch {
		case project.Name == "":
			continue
		case project.Title == "":
			continue
		default:
			valid = append(valid, project)
		}
	}

	allowed := make([]Project, 0, len(valid))
	for _, project := range valid {
		if project.AccessibleBy(identity) {
			allowed = append(allowed, project)
		}
	}
	return allowed[:], nil
}

func GetProject(projects []Project, name string) (Project, bool) {
	for _, project := range projects {
		if project.Name == name {
			return project, true
		}
	}
	return Project{}, false
}

func (x Project) AccessibleBy(accessor auth.Identity) bool {
	switch {
	case x.PartsReleased == true:
		return true
	case accessor.HasRole(auth.RoleVVGOProductionTeam):
		return true
	case accessor.HasRole(auth.RoleVVGOExecutiveDirector):
		return true
	default:
		return false
	}
}

func (x Project) ProjectPage() string { return "/projects?name=" + x.Name }
func (x Project) PartsPage() string   { return "/parts?project=" + x.Name }
func (x Project) IsAcceptingSubmissions() bool {
	switch {
	case x.Hidden:
		return false
	case x.VideoReleased:
		return false
	case x.PartsArchived:
		return false
	default:
		return x.PartsReleased
	}
}
