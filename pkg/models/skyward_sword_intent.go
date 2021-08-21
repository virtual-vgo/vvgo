package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/sheets"
	"sort"
)

const SkywardSwordIntentID = "1X5TKohjuEJS8M5wwOOArS8SP9lU7TsUbKbk6KZjddFw"

type SkywardSwordIntent struct {
	CreditedName   string `col_name:"Credited Name"`
	DiscordHandle  string `col_name:"Discord Handle"`
	IntendToRecord string `col_name:"Part You Intend to Record"`
	OtherPart      string `col_name:"If you could play another part, what would you play?"`
}

type SkywardSwordIntents []SkywardSwordIntent

func ListSkywardSwordIntents(ctx context.Context) (SkywardSwordIntents, error) {
	values, err := sheets.ReadSheet(ctx, SkywardSwordIntentID, "Form Responses 1")
	if err != nil {
		return nil, err
	}
	return valuesToSkywardSwordIntents(values), nil
}

func valuesToSkywardSwordIntents(values [][]interface{}) []SkywardSwordIntent {
	if len(values) < 1 {
		return nil
	}
	index := sheets.BuildIndex(values[0])
	docs := make([]SkywardSwordIntent, len(values)-1)
	for i, row := range values[1:] {
		sheets.ProcessRow(row, &docs[i], index)
	}
	SkywardSwordIntents(docs).Sort()
	return docs
}

func (x SkywardSwordIntents) Len() int           { return len(x) }
func (x SkywardSwordIntents) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x SkywardSwordIntents) Less(i, j int) bool { return x[i].IntendToRecord < x[j].IntendToRecord }
func (x SkywardSwordIntents) Sort()              { sort.Sort(x) }
