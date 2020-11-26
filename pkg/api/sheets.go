package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"google.golang.org/api/sheets/v4"
	"reflect"
	"sort"
	"strconv"
)

const (
	ProjectsRange = "Projects"
	PartsRange    = "Parts"
	CreditsRange  = "Credits"
	LeadersRange  = "Leaders"
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

func ValuesToProjects(values [][]interface{}) []Project {
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

func ValuesToParts(values [][]interface{}) []Part {
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

func ValuesToCredits(values [][]interface{}) []Credit {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	credits := make([]Credit, len(values)-1)
	for i, row := range values[1:] {
		processRow(row, &credits[i], index)
	}
	CreditsSort(credits).Sort()
	return credits
}

func ValuesToLeaders(values [][]interface{}) []Leader {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	leaders := make([]Leader, len(values)-1)
	for i, row := range values[1:] {
		processRow(row, &leaders[i], index)
	}
	return leaders
}

func buildIndex(fieldNames []interface{}) map[string]int {
	index := make(map[string]int, len(fieldNames))
	for i, col := range fieldNames {
		index[fmt.Sprintf("%s", col)] = i
	}
	return index
}

func readValuesFromRedis(ctx context.Context, spreadsheetID string, readRange string) ([][]interface{}, error) {
	var buf bytes.Buffer
	key := "sheets:" + spreadsheetID + ":" + readRange
	if err := redis.Do(ctx, redis.Cmd(&buf, "GET", key)); err != nil {
		return nil, err
	}
	var values [][]interface{}
	if err := json.NewDecoder(&buf).Decode(&values); err != nil {
		return nil, err
	}
	return values, nil
}

func writeValuesToRedis(ctx context.Context, spreadsheetID string, readRange string, values [][]interface{}) error {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&values)
	key := "sheets:" + spreadsheetID + ":" + readRange
	return redis.Do(ctx, redis.Cmd(nil, "SETEX", key, "1", buf.String()))
}

func readSheet(ctx context.Context, spreadsheetID string, readRange string) ([][]interface{}, error) {

	// first try to read from redis
	values, err := readValuesFromRedis(ctx, spreadsheetID, readRange)
	if err != nil {
		logger.WithError(err).Infof("failed to read spreadsheet values from redis")
	} else if len(values) != 0 {
		return values, nil
	}

	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, fmt.Errorf("no data")
	}

	if err := writeValuesToRedis(ctx, spreadsheetID, readRange, resp.Values); err != nil {
		logger.WithError(err).Errorf("failed to write spreadsheet values to redis")
	}

	return resp.Values, nil
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
