package api

import (
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"reflect"
	"sort"
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

type Credit struct {
	Project       string
	Order         int
	MajorCategory string `col_name:"Major Category"`
	MinorCategory string `col_name:"Minor Category"`
	Name          string
	BottomText    string `col_name:"Bottom Text"`
}

type CreditsSort []Credit

func (x CreditsSort) Len() int           { return len(x) }
func (x CreditsSort) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x CreditsSort) Less(i, j int) bool { return x[i].Order < x[j].Order }
func (x CreditsSort) Sort()              { sort.Sort(x) }

type Leader struct {
	Name         string
	Epithet      string
	Affiliations string
	Blurb        string
	Icon         string
	Email        string
}

func listProjects(ctx context.Context, spreadsheetID string) ([]Project, error) {
	resp, index, err := readSheet(ctx, spreadsheetID, "Projects")
	if err != nil {
		return nil, err
	}

	projects := make([]Project, len(resp.Values)-1) // ignore the header row
	for i, row := range resp.Values[1:] {
		processRow(row, &projects[i], index)
	}
	return projects, nil
}

func listParts(ctx context.Context, spreadsheetID string) ([]Part, error) {
	resp, index, err := readSheet(ctx, spreadsheetID, "Parts")
	if err != nil {
		return nil, err
	}

	parts := make([]Part, len(resp.Values)-1)
	for i, row := range resp.Values[1:] {
		processRow(row, &parts[i], index)
	}
	return parts, nil
}

func listCredits(ctx context.Context, spreadsheetID string) ([]Credit, error) {
	resp, index, err := readSheet(ctx, spreadsheetID, "Credits")
	if err != nil {
		return nil, err
	}

	credits := make([]Credit, len(resp.Values)-1)
	for i, row := range resp.Values[1:] {
		processRow(row, &credits[i], index)
	}
	return credits, nil
}

func listLeaders(ctx context.Context, spreadsheetID string) ([]Leader, error) {
	resp, index, err := readSheet(ctx, spreadsheetID, "Leaders")
	if err != nil {
		return nil, err
	}

	leaders := make([]Leader, len(resp.Values)-1)
	for i, row := range resp.Values[1:] {
		processRow(row, &leaders[i], index)
	}
	return leaders, nil
}

func readSheet(ctx context.Context, spreadsheetID string, readRange string) (*sheets.ValueRange, map[string]int, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, nil, fmt.Errorf("no data")
	}

	index := make(map[string]int, len(resp.Values[0])-1)
	for i, col := range resp.Values[0] {
		index[fmt.Sprintf("%s", col)] = i
	}

	return resp, index, nil
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
