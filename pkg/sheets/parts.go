package sheets

import (
	"context"
)

type Part struct {
	Project            string
	PartName           string `col_name:"Part Name"`
	ScoreOrder         int    `col_name:"Score Order"`
	SheetMusicFile     string `col_name:"Sheet Music File"`
	ClickTrackFile     string `col_name:"Click Track File"`
	ConductorVideo     string `col_name:"Conductor Video"`
	ReferenceTrack     string `col_name:"Reference Track"`
	PronunciationGuide string `col_name:"Pronunciation Guide"`
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
	}
	return parts
}

func (x Parts) ForProject(project string) Parts {
	var want Parts
	for _, part := range x {
		if part.Project == project {
			want = append(want, part)
		}
	}
	return want
}

func (x Parts) Append(parts Parts) Parts {
	return append(x, parts...)
}
