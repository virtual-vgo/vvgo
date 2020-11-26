package leader

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
)

type Leader struct {
	Name         string
	Epithet      string
	Affiliations string
	Blurb        string
	Icon         string
	Email        string
}

type Leaders []Leader

func List(ctx context.Context, spreadsheetID string) (Leaders, error) {
	values, err := sheets.ReadSheet(ctx, spreadsheetID, "Leaders")
	if err != nil {
		return nil, err
	}
	return valuesToLeaders(values), nil
}

func valuesToLeaders(values [][]interface{}) Leaders {
	if len(values) < 1 {
		return nil
	}
	index := sheets.BuildIndex(values[0])
	leaders := make([]Leader, len(values)-1)
	for i, row := range values[1:] {
		sheets.ProcessRow(row, &leaders[i], index)
	}
	return leaders
}
