package api

import (
	"fmt"
	"strings"
)

type SubmissionRecord struct {
	CreditedName string `col_name:"Credited Name"`
	Instrument   string
	BottomText   string `col_name:"Bottom Text"`
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
		} else {
			credit.BottomText += ", " + record.BottomText
		}
		creditsMap[record.Instrument+record.CreditedName] = credit
	}
	credits := make([]Credit, 0, len(creditsMap))
	for _, credit := range creditsMap {
		credit.BottomText += ")"
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
		output += fmt.Sprintf("%s\t%d\t%s\t%s\t%s\t%s\n", credit.Project, credit.Order, credit.MajorCategory,
			credit.MinorCategory, credit.Name, credit.BottomText)
	}
	return output
}

func CreditsToVideoPasta(credits []Credit) string {
	if len(credits) == 0 {
		return "\n"
	}

	var output string
	var lastMinor = credits[0].MinorCategory
	for _, credit := range credits {
		if credit.MinorCategory != lastMinor {
			output += fmt.Sprintf("\n%s\t%s", credit.MinorCategory,
				strings.ReplaceAll(credit.MinorCategory, "♭", "_"))
		}
		if  credit.BottomText == "" {
			output += fmt.Sprintf("\t%s", credit.Name)
		} else {
			output += fmt.Sprintf("\t%s %s", credit.Name, credit.BottomText)
		}
	}
	return output + "\n"
}

func CreditsToYoutubePasta(credits []Credit) string {
	if len(credits) == 0 {
		return "\n"
	}

	var output string
	var lastMinor = credits[0].MinorCategory
	for _, credit := range credits {
		if credit.MinorCategory != lastMinor {
			output += fmt.Sprintf("\n%s", credit.MinorCategory)
		}
		if  credit.BottomText == "" {
			output += fmt.Sprintf(", %s", credit.Name)
		} else {
			output += fmt.Sprintf(", %s %s", credit.Name, credit.BottomText)
		}	}
	return output + "\n"
}
