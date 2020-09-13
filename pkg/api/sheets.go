package api

import (
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"reflect"
	"strconv"
)

type Project struct {
	Name                    string
	Title                   string
	Released                bool
	Archived                bool
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
	YoutubeLink             string `col_name:"Youtube Link"`
	YoutubeEmbed            string `col_name:"Youtube Embed"`
	SubmissionDeadline      string `col_name:"Submission Deadline"`
	SubmissionLink          string `col_name:"Submission Link"`
	Season                  string
	BannerLink              string `col_name:"Banner Link"`
}

func listProjects(ctx context.Context, spreadSheetID string) ([]Project, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Projects"
	resp, err := srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, fmt.Errorf("no data")
	}
	projects := make([]Project, len(resp.Values)-1) // ignore the header row

	index := make(map[string]int, len(resp.Values[0])-1)
	for i, col := range resp.Values[0] {
		index[fmt.Sprintf("%s", col)] = i
	}

	for i, row := range resp.Values[1:] {
		processRow(row, &projects[i], index)
	}
	return projects, nil
}

type Part struct {
	Project            string
	ProjectTitle       string `col_name:"Project Title"`
	PartName           string `col_name:"Part Name"`
	ScoreOrder         int    `col_name:"Score Order"`
	SheetMusicFile     string `col_name:"Sheet Music File"`
	ClickTrackFile     string `col_name:"Click Track File"`
	ConductorVideo     string `col_name:"Conductor Video"`
	Released           bool
	Archived           bool
	ReferenceTrack     string `col_name:"Reference Track"`
	PronunciationGuide string `col_name:"Pronunciation Guide"`
}

func listParts(ctx context.Context, spreadSheetID string) ([]Part, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Parts"
	resp, err := srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, fmt.Errorf("no data")
	}
	parts := make([]Part, len(resp.Values)-1)

	index := make(map[string]int, len(resp.Values[0])-1)
	for i, col := range resp.Values[0] {
		index[fmt.Sprintf("%s", col)] = i
	}

	for i, row := range resp.Values[1:] {
		processRow(row, &parts[i], index)
	}
	return parts, nil
}

func processRow(row []interface{}, dest interface{}, index map[string]int) {
	tagName := "col_name"
	if len(row) < 1 {
		return
	}
	reflectType := reflect.TypeOf(dest).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		colName := field.Tag.Get(tagName)
		if colName == "" {
			colName = field.Name
		}
		colIndex, ok := index[colName]
		if !ok {
			continue
		}
		if len(row) > colIndex {
			switch field.Type.Kind() {
			case reflect.String:
				val := fmt.Sprint(row[colIndex])
				reflect.ValueOf(dest).Elem().Field(i).SetString(val)
			case reflect.Bool:
				val, _ := strconv.ParseBool(fmt.Sprint(row[colIndex]))
				reflect.ValueOf(dest).Elem().Field(i).SetBool(val)
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				val, _ := strconv.ParseInt(fmt.Sprint(row[colIndex]), 10, 64)
				reflect.ValueOf(dest).Elem().Field(i).SetInt(val)
			}
		}
	}
}
