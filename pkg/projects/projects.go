package projects

import (
	"github.com/virtual-vgo/vvgo/pkg/log"
)

// A VVGO project
type Project struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Released     bool     `json:"released"`
	Archived     bool     `json:"archived"`
	Sources      []string `json:"sources"`
	Composers    []string `json:"composers"`
	Arrangers    []string `json:"arrangers"`
	Editors      []string `json:"editors"`
	Transcribers []string `json:"transcribers"`
	Preparers    []string `json:"preparers"`
	ClixBy       []string `json:"clix_by"`
	Reviewers    []string `json:"reviewers"`
	Lyricists    []string `json:"lyricists"`
	AddlContent  []string `json:"addl_content"`
}

var logger = log.Logger()

var project = Projects{projects: []Project{
	{
		Name:         "01-snake-eater",
		Title:        "Snake Eater",
		Released:     true,
		Archived:     false,
		Sources:      []string{"Metal Gear Solid 3"},
		Composers:    []string{},
		Arrangers:    []string{},
		Editors:      []string{"Jerome Landingin"},
		Transcribers: []string{},
		Preparers:    []string{},
		ClixBy:       []string{},
		Reviewers:    []string{},
		Lyricists:    []string{"B. Harnish"},
		AddlContent:  []string{},
	},
}}

func init() {
	// build indices
	project.nameIndex = make(Index)
	for i, p := range project.projects {
		project.nameIndex[p.Name] = &project.projects[i]
	}
}

func GetName(name string) *Project { return project.GetName(name) }
func Exists(name string) bool      { return project.Exists(name) }

type Projects struct {
	projects  []Project
	nameIndex Index
}

type Index map[string]*Project

func (x *Projects) GetName(name string) *Project {
	return x.nameIndex[name]
}

func (x *Projects) Exists(name string) bool {
	_, ok := x.nameIndex[name]
	return ok
}
