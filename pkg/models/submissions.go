package models

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/sheets"
	"strings"
)

type Submission struct {
	CreditedName string
	Instrument   string
	BottomText   string
}

type Submissions []Submission

func ListSubmissions(ctx context.Context, spreadsheetID string, readRange string) (Submissions, error) {
	values, err := sheets.ReadSheet(ctx, spreadsheetID, readRange)
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
	submissionRecords := make([]Submission, 0, len(values)-1) // ignore the header row
	UnmarshalSheet(values, &submissionRecords)
	return submissionRecords
}

func (x Submissions) ToCredits(projectName string) Credits {
	creditsMap := make(map[string]*Credit)
	for i, record := range x {
		submissionCredit := creditsMap[record.Instrument+record.CreditedName]
		if submissionCredit == nil {
			submissionCredit = &Credit{
				Project:       projectName,
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
