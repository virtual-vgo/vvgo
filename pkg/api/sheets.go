package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/sheet"
	"net/http"
	"path/filepath"
)

func (x *Server) SheetsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		sheet.Sheet
		Link string `json:"link"`
	}

	sheetsStorage := &sheet.Storage{
		RedisLocker: x.RedisLocker,
		MinioDriver: x.MinioDriver,
	}

	allSheets := sheetsStorage.List()
	rows := make([]tableRow, 0, len(allSheets))
	for _, sheet := range allSheets {
		rows = append(rows, tableRow{
			Sheet: sheet,
			Link:  sheet.Link(),
		})
	}

	var buffer bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(&buffer, &rows, filepath.Join(Public, "sheets.gohtml")); !ok {
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
