package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

type CreditsMaker struct{}

func (x CreditsMaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data := struct {
		SpreadsheetID string
		ReadRange     string
		Project       string
		ErrorMessage  string
		WebsitePasta  string
		VideoPasta    string
		YoutubePasta  string
	}{
		SpreadsheetID: r.FormValue("spreadsheetID"),
		ReadRange:     r.FormValue("readRange"),
		Project:       r.FormValue("project"),
	}

	if data.SpreadsheetID != "" && data.ReadRange != "" {
		values, err := readSheet(ctx, data.SpreadsheetID, data.ReadRange)
		if err != nil {
			logger.WithError(err).Error("readSheet() failed")
			data.ErrorMessage = err.Error()
		} else {
			records := ValuesToSubmissionRecords(values)
			credits := SubmissionRecordsToCredits(data.Project, records)
			data.WebsitePasta = CreditsToWebsitePasta(credits)
			data.VideoPasta = CreditsToVideoPasta(credits)
			data.YoutubePasta = CreditsToYoutubePasta(credits)
		}
	}

	// set defaults
	if data.SpreadsheetID == "" {
		data.SpreadsheetID = "1BP3fGC2C6mKe3ZuVhby4eCxidlHL768bDdHsJ5mQleo"
	}
	if data.ReadRange == "" {
		data.ReadRange = "06 Aurene!A3:I39"
	}
	if data.Project == "" {
		data.Project = "06-aurene-dragon-full-of-light"
	}
	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &data, "credits-maker.gohtml"); !ok {
		internalServerError(w)
		return
	}
	body, _ := httputil.DumpRequest(r, true)
	fmt.Println(string(body))
	_, _ = buffer.WriteTo(w)
}

type SubmissionRecord struct {
	CreditedName string `col_name:"Credited Name"`
	Instrument   string
	BottomText   string `col_name:"Bottom Text"`
}

func ValuesToSubmissionRecords(values [][]interface{}) []SubmissionRecord {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	submissionRecords := make([]SubmissionRecord, len(values)-1) // ignore the header row
	for i, row := range values[1:] {
		processRow(row, &submissionRecords[i], index)
	}
	return submissionRecords
}

func SubmissionRecordsToCredits(project string, records []SubmissionRecord) []Credit {
	creditsMap := make(map[string]*Credit)
	for i, record := range records {
		credit := creditsMap[record.Instrument+record.CreditedName]
		if credit == nil {
			credit = &Credit{
				Project:       project,
				Order:         i,
				MajorCategory: "PERFORMERS",
				MinorCategory: strings.ToUpper(record.Instrument),
				Name:          record.CreditedName,
				BottomText:    "(" + record.BottomText,
			}
		} else if record.BottomText != "" {
			credit.BottomText += ", " + record.BottomText
		}
		creditsMap[record.Instrument+record.CreditedName] = credit
	}
	credits := make([]Credit, 0, len(creditsMap))
	for _, credit := range creditsMap {
		credit.BottomText += ")"
		credit.BottomText = strings.ToUpper(credit.BottomText)
		if credit.BottomText == "()" {
			credit.BottomText = ""
		}
		credits = append(credits, *credit)
	}
	CreditsSort(credits).Sort()
	return credits
}

func CreditsToWebsitePasta(credits []Credit) string {
	var output string
	for _, credit := range credits {
		output += strings.TrimSpace(fmt.Sprintf("%s\t\t%s\t%s\t%s\t%s", credit.Project, credit.MajorCategory,
			credit.MinorCategory, credit.Name, credit.BottomText)) + "\n"
	}
	return output
}

func CreditsToVideoPasta(credits []Credit) string {
	output := "— PERFORMERS —\t— PERFORMERS —"
	if len(credits) == 0 {
		return output
	}
	var lastMinor string
	for _, credit := range credits {
		if credit.MinorCategory != lastMinor {
			lastMinor = credit.MinorCategory
			output += fmt.Sprintf("\n%s\t%s", credit.MinorCategory,
				strings.ReplaceAll(credit.MinorCategory, "♭", "_"))
		}
		if credit.BottomText == "" {
			output += fmt.Sprintf("\t%s", credit.Name)
		} else {
			output += fmt.Sprintf("\t%s %s", credit.Name, credit.BottomText)
		}
	}
	return output + "\n"
}

func CreditsToYoutubePasta(credits []Credit) string {
	output := "— PERFORMERS —"
	if len(credits) == 0 {
		return output
	}
	var lastMinor string
	for _, credit := range credits {
		if credit.MinorCategory != lastMinor {
			lastMinor = credit.MinorCategory
			output += fmt.Sprintf("\n\n%s\n", credit.MinorCategory)
		} else {
			output += ", "
		}
		if credit.BottomText == "" {
			output += fmt.Sprintf("%s", credit.Name)
		} else {
			output += fmt.Sprintf("%s %s", credit.Name, credit.BottomText)
		}
	}
	return output + "\n"
}
