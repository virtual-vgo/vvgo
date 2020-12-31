package sheets

import (
	"context"
	"sort"
)

type Part struct {
	Project            string
	PartName           string `col_name:"Part Name"`
	ScoreOrder         int    `col_name:"Score Order"`
	SheetMusicFile     string `col_name:"Sheet Music File"`
	ClickTrackFile     string `col_name:"Click Track File"`
	ConductorVideo     string `col_name:"Conductor Video"`
	PronunciationGuide string `col_name:"Pronunciation Guide"`

	// Derived Columns
	SheetMusicLink         string
	ClickTrackLink         string
	PronunciationGuideLink string
}

type Parts []Part

func ListParts(ctx context.Context) (Parts, error) {
	values, err := ReadSheet(ctx, WebsiteDataSpreadsheetID(ctx), "Parts")
	if err != nil {
		return nil, err
	}
	return valuesToParts(values), nil
}

func valuesToParts(values [][]interface{}) Parts {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	parts := make([]Part, len(values)-1)
	for i, row := range values[1:] {
		processRow(row, &parts[i], index)
		parts[i].SheetMusicLink = downloadLink(parts[i].SheetMusicFile)
		parts[i].ClickTrackLink = downloadLink(parts[i].ClickTrackFile)
		parts[i].PronunciationGuideLink = downloadLink(parts[i].PronunciationGuideLink)
	}
	return parts
}

func downloadLink(object string) string {
	if object == "" {
		return ""
	} else {
		return "/download?object=" + object
	}
}

func (x Parts) ForProject(projects ...string) Parts {
	var want Parts
	for _, part := range x {
		for _, project := range projects {
			if part.Project == project {
				want = append(want, part)
			}
		}
	}
	return want
}

func (x Parts) Append(parts Parts) Parts {
	return append(x, parts...)
}

// Sorting

func (x Parts) Len() int           { return len(x) }
func (x Parts) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x Parts) Less(i, j int) bool { return x[i].ScoreOrder < x[j].ScoreOrder }
func (x Parts) Sort() Parts        { sort.Sort(x); return x }
