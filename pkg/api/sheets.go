package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"path/filepath"
)

func (x *Server) SheetsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	type tableRow struct {
		sheets.Sheet
		Link string `json:"link"`
	}

	allSheets := x.sheetsStorage.List()
	rows := make([]tableRow, 0, len(allSheets))
	for _, sheet := range allSheets {
		rows = append(rows, tableRow{
			Sheet: sheet,
			Link:  sheet.Link(),
		})
	}

	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(w, &rows, filepath.Join(Public, "sheets.gohtml")); !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}
	default:
		if ok := jsonEncode(w, &rows); !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}
