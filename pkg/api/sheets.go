package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/sheet"
	"net/http"
	"path/filepath"
)

const SheetsBucketName = "sheets"
const SheetsLockerKey = "sheets.lock"

type SheetsHandler struct {
	sheet.Sheets
}

func (x SheetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		sheet.Sheet
		Link string `json:"link"`
	}

	sheets := x.List()
	rows := make([]tableRow, 0, len(sheets))
	for _, sheet := range sheets {
		rows = append(rows, tableRow{
			Sheet: sheet,
			Link:  sheet.Link(SheetsBucketName),
		})
	}

	var buffer bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(&buffer, &rows, filepath.Join(PublicFiles, "sheets.gohtml")); !ok {
			internalServerError(w)
			return
		}
	default:
		if ok := jsonEncode(&buffer, &rows); !ok {
			internalServerError(w)
			return
		}
	}
	buffer.WriteTo(w)
}
