package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"path/filepath"
)

const SheetsBucketName = "sheets"
const SheetsLockerKey = "sheets.lock"

type SheetsHandler struct {
	sheets.Sheets
}

func (x SheetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		sheets.Sheet
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
		jsonEncode(&buffer, &rows)
	}
	buffer.WriteTo(w)
}
