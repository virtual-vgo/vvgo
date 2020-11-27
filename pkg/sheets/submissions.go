package sheets

import (
	"context"
	"fmt"
	"strings"
)

type Submission struct {
	CreditedName string `col_name:"Credited Name"`
	Instrument   string
	BottomText   string `col_name:"Bottom Text"`
}

type Submissions []Submission

func ListSubmissions(ctx context.Context, spreadsheetID string, readRange string) (Submissions, error) {
	values, err := ReadSheet(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%#v\n", values)
	return valuesToSubmissionRecords(values), nil
}

func valuesToSubmissionRecords(values [][]interface{}) Submissions {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	submissionRecords := make([]Submission, len(values)-1) // ignore the header row
	for i, row := range values[1:] {
		processRow(row, &submissionRecords[i], index)
	}
	return submissionRecords
}

func (x Submissions) ToCredits(project string) Credits {
	creditsMap := make(map[string]*Credit)
	for i, record := range x {
		submissionCredit := creditsMap[record.Instrument+record.CreditedName]
		if submissionCredit == nil {
			submissionCredit = &Credit{
				Project:       project,
				Order:         i,
				MajorCategory: "PERFORMERS",
				MinorCategory: strings.ToUpper(record.Instrument),
				Name:          record.CreditedName,
				BottomText:    "(" + record.BottomText,
			}
		} else if record.BottomText != "" {
			submissionCredit.BottomText += ", " + record.BottomText
		}
		creditsMap[record.Instrument+record.CreditedName] = submissionCredit
	}
	credits := make(Credits, 0, len(creditsMap))
	for _, submissionCredit := range creditsMap {
		submissionCredit.BottomText += ")"
		submissionCredit.BottomText = strings.ToUpper(submissionCredit.BottomText)
		if submissionCredit.BottomText == "()" {
			submissionCredit.BottomText = ""
		}
		credits = append(credits, *submissionCredit)
	}
	credits.Sort()
	return credits
}
