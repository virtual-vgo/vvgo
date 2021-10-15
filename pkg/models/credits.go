package models

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"sort"
	"strings"
)

const SheetCredits = "Credits"

type Credit struct {
	Project       string
	Order         int
	MajorCategory string `col_name:"Major Category"`
	MinorCategory string `col_name:"Minor Category"`
	Name          string
	BottomText    string `col_name:"Bottom Text"`
}

type Credits []Credit

func ListCredits(ctx context.Context) (Credits, error) {
	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetCredits)
	if err != nil {
		return nil, err
	}
	return valuesToCredits(values), nil
}

func valuesToCredits(values [][]interface{}) []Credit {
	if len(values) < 1 {
		return nil
	}
	credits := make([]Credit, 0, len(values)-1)
	UnmarshalSheet(values, &credits)
	Credits(credits).Sort()
	return credits
}

func (x Credits) Len() int           { return len(x) }
func (x Credits) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x Credits) Less(i, j int) bool { return x[i].Order < x[j].Order }
func (x Credits) Sort()              { sort.Sort(x) }

func (x Credits) ForProject(name string) Credits {
	var want Credits
	for _, credit := range x {
		if credit.Project == name {
			want = append(want, credit)
		}
	}
	return want
}

func (x Credits) WebsitePasta() string {
	var output string
	for _, credit := range x {
		output += strings.TrimSpace(fmt.Sprintf("%s\t\t%s\t%s\t%s\t%s", credit.Project, credit.MajorCategory,
			credit.MinorCategory, credit.Name, credit.BottomText)) + "\n"
	}
	return output
}

func (x Credits) VideoPasta() string {
	output := "— PERFORMERS —\t— PERFORMERS —"
	var lastMinor string
	for _, credit := range x {
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

func (x Credits) YoutubePasta() string {
	output := "— PERFORMERS —"
	var lastMinor string
	for _, credit := range x {
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
