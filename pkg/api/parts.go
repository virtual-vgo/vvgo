package api

import (
	"bytes"
	"net/http"
	"path/filepath"
)

type PartsHandler struct{ *Database }

func (x PartsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		Project    string `json:"project"`
		PartName   string `json:"part_name"`
		PartNumber uint8  `json:"part_number"`
		SheetMusic string `json:"sheet_music"`
		ClickTrack string `json:"click_track"`
	}

	allParts := x.Parts.List()
	rows := make([]tableRow, 0, len(allParts))
	for _, part := range allParts {
		rows = append(rows, tableRow{
			Project:    part.Project,
			PartName:   part.Name,
			PartNumber: part.Number,
			SheetMusic: part.SheetLink(SheetsBucketName),
			ClickTrack: part.ClickLink(ClixBucketName),
		})
	}

	var buffer bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(&buffer, &rows, filepath.Join(PublicFiles, "parts.gohtml")); !ok {
			internalServerError(w)
			return
		}
	default:
		jsonEncode(&buffer, &rows)
	}
	buffer.WriteTo(w)
}
