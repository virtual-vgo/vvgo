package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"sort"
)

const SheetParts = "Parts"

type Part struct {
	Project            string `json:"project"`
	PartName           string `json:"part_name"`
	ScoreOrder         int    `json:"score_order"`
	SheetMusicFile     string `json:"sheet_music_file"`
	ClickTrackFile     string `json:"click_track_file"`
	ConductorVideo     string `json:"conductor_video"`
	PronunciationGuide string `json:"pronunciation_guide"`

	// Derived Columns
	SheetMusicLink         string `json:"sheet_music_link"`
	ClickTrackLink         string `json:"click_track_link"`
	PronunciationGuideLink string `json:"pronunciation_guide_link"`
}

type Parts []Part

func ListParts(ctx context.Context) (Parts, error) {
	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetParts)
	if err != nil {
		return nil, err
	}
	return valuesToParts(values), nil
}

func valuesToParts(values [][]interface{}) Parts {
	if len(values) < 1 {
		return nil
	}
	parts := make([]Part, 0, len(values)-1)
	UnmarshalSheet(values, &parts)
	for i := range parts {
		if parts[i].SheetMusicLink == "" {
			parts[i].SheetMusicLink = downloadLink(parts[i].SheetMusicFile)
		}
		if parts[i].ClickTrackLink == "" {
			parts[i].ClickTrackLink = downloadLink(parts[i].ClickTrackFile)
		}
		if parts[i].PronunciationGuideLink == "" {
			parts[i].PronunciationGuideLink = downloadLink(parts[i].PronunciationGuide)
		}
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
