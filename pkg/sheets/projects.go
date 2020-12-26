package sheets

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"sort"
)

type Project struct {
	Name                    string
	Title                   string
	Season                  string
	Hidden                  bool
	PartsReleased           bool `col_name:"Parts Released"`
	PartsArchived           bool `col_name:"Parts Archived"`
	VideoReleased           bool `col_name:"Video Released"`
	Sources                 string
	Composers               string
	Arrangers               string
	Editors                 string
	Transcribers            string
	Preparers               string
	ClixBy                  string `col_name:"Clix By"`
	Reviewers               string
	Lyricists               string
	AdditionalContent       string `col_name:"Additional Content"`
	ReferenceTrack          string `col_name:"Reference Track"`
	ChoirPronunciationGuide string `col_name:"Choir Pronunciation Guide"`
	BannerLink              string `col_name:"Banner Link"`
	YoutubeLink             string `col_name:"Youtube Link"`
	YoutubeEmbed            string `col_name:"Youtube Embed"`
	SubmissionDeadline      string `col_name:"Submission Deadline"`
	SubmissionLink          string `col_name:"Submission Link"`
}

func (x Project) ProjectPage() string { return "/projects/" + x.Name }
func (x Project) PartsPage() string   { return "/parts?project=" + x.Name }

type Projects []Project

func ListProjects(ctx context.Context, identity *login.Identity) (Projects, error) {
	values, err := ReadSheet(ctx, WebsiteDataSpreadsheetID(ctx), "Projects")
	if err != nil {
		return nil, err
	}
	return valuesToProjects(values).ForIdentity(identity), nil
}

func valuesToProjects(values [][]interface{}) Projects {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	projects := make([]Project, len(values)-1) // ignore the header row
	for i, row := range values[1:] {
		processRow(row, &projects[i], index)
	}
	return projects
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

func (x Projects) ForIdentity(identity *login.Identity) Projects {
	var want Projects
	for _, project := range x {
		switch {
		case project.PartsReleased == true:
			want = append(want, project)
		case identity.HasRole(login.RoleVVGOTeams):
			want = append(want, project)
		case identity.HasRole(login.RoleVVGOLeader):
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

// Sorting

func (x Projects) Len() int              { return len(x) }
func (x Projects) Swap(i, j int)         { x[i], x[j] = x[j], x[i] }
func (x Projects) Less(i, j int) bool    { return x[i].Name < x[j].Name }
func (x Projects) Sort() Projects        { sort.Sort(x); return x }
func (x Projects) ReverseSort() Projects { sort.Sort(sort.Reverse(x)); return x }
