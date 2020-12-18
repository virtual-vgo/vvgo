package sheets

import (
	"context"
)

type Leaders []Leader

type Leader struct {
	Name         string
	Epithet      string
	Affiliations string
	Blurb        string
	Icon         string
	Email        string
}

func ListLeaders(ctx context.Context) (Leaders, error) {
	values, err := ReadSheet(ctx, WebsiteDataSpreadsheetID(ctx), "Leaders")
	if err != nil {
		return nil, err
	}
	return valuesToLeaders(values), nil
}

func valuesToLeaders(values [][]interface{}) Leaders {
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
