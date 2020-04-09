package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"path/filepath"
)

func (x *Server) SheetsIndex(w http.ResponseWriter, r *http.Request) {
	// only accept get
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	type tableRow struct {
		sheets.Sheet
		Link string `json:"link"`
	}

	objects := x.ListObjects(sheets.BucketName)
	rows := make([]tableRow, 0, len(objects))
	for i := range objects {
		sheet := sheets.NewSheetFromTags(objects[i].Tags)
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
