package models

import (
	"context"
	"sort"

	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
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
	Mixtape                 bool
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

func (x Project) ProjectPage() string { return "/projects?name=" + x.Name }
func (x Project) PartsPage() string   { return "/parts?project=" + x.Name }

type Projects []Project

func ListProjects(ctx context.Context, identity Identity) (Projects, error) {
	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetProjects)
	if err != nil {
		return nil, err
	}
	return ValuesToProjects(values).ForIdentity(identity), nil
}

func ValuesToProjects(values [][]interface{}) Projects {
	if len(values) < 1 {
		return nil
	}
	var projects Projects
	UnmarshalSheet(values, &projects)
	return projects.prune()
}

func (x Projects) prune() Projects {
	want := make(Projects, 0, len(x))
	for _, project := range x {
		switch {
		case project.Name == "":
			continue
		case project.Title == "":
			continue
		default:
			want = append(want, project)
		}
	}
	return want[:]
}

func (x Projects) Current() Projects {
	return x.Query(map[string]interface{}{
		"Hidden": false, "Video Released": false,
		"Parts Archived": false, "Parts Released": true})
}

func (x Projects) Get(name string) (Project, bool) {
	for _, project := range x {
		if project.Name == name {
			return project, true
		}
	}
	return Project{}, false
}

func (x Projects) Has(name string) bool { _, ok := x.Get(name); return ok }

func (x Projects) ForIdentity(identity Identity) Projects {
	var want Projects
	for _, project := range x {
		switch {
		case project.PartsReleased == true:
			want = append(want, project)
		case identity.HasRole(RoleVVGOProductionTeam):
			want = append(want, project)
		case identity.HasRole(RoleVVGOExecutiveDirector):
			want = append(want, project)
		}
	}
	return want
}

func (x Projects) Select(names ...string) Projects {
	var want Projects
	for _, name := range names {
		project, ok := x.Get(name)
		if ok {
			want = append(want, project)
		}
	}
	return want
}

func (x Projects) Append(projects Projects) Projects { return append(x, projects...) }

func (x Projects) First() Project {
	if x.Len() == 0 {
		return Project{}
	} else {
		return x[0]
	}
}

func (x Projects) Last() Project {
	if x.Len() == 0 {
		return Project{}
	} else {
		return x[x.Len()-1]
	}
}

func (x Projects) WithField(key string, value interface{}) Projects {
	return x.Query(NewQuery().WithField(key, value))
}

func (x Projects) WithFields(args ...interface{}) Projects {
	return x.Query(NewQuery().WithFieldsUnsafe(args...))
}

func (x Projects) Query(query map[string]interface{}) Projects {
	want := make(Projects, 0, x.Len())
	Query(query).MatchSlice(x, &want)
	return want[:]
}

func (x Projects) Names() []string {
	names := make([]string, x.Len())
	for i := range x {
		names[i] = x[i].Name
	}
	return names
}

// Sorting

func (x Projects) Len() int              { return len(x) }
func (x Projects) Swap(i, j int)         { x[i], x[j] = x[j], x[i] }
func (x Projects) Less(i, j int) bool    { return x[i].Name < x[j].Name }
func (x Projects) Sort() Projects        { sort.Sort(x); return x }
func (x Projects) ReverseSort() Projects { sort.Sort(sort.Reverse(x)); return x }
