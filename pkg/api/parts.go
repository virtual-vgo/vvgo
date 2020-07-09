package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"io"
	"net/http"
	"strconv"
	"time"
)

type PartsUpdateHandler struct {
	database *Database
}

func (x PartsUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "parts_update_handler")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		// render upload form
	case http.MethodPost:
		// parse and validate file and post parameters
		file, fileHeader, err := r.FormFile("partsFile")
		if err != nil {
			logger.WithError(err).Error("r.FomFile() failed")
			badRequest(w, "invalid form")
			return
		}
		defer file.Close()

		// check the content type
		var upload []parts.Part
		switch contentType := fileHeader.Header.Get("Content-Type"); contentType {
		case "application/json"
			json.NewDecoder(r.Body).Decode(&upload)
		case "application/csv":
			upload = partsFromCSV(r.Body)
		default:
			badRequest(w, fmt.Sprintf("invalid content type: %s", contentType))
		}
		if err := x.database.Parts.Save(ctx, upload); err != nil {
			logger.WithError(err).Error("x.database.Parts.Save() failed")
			internalServerError(w)
			return
		}

	default:
		badRequest(w, "unsupported method")
	}
}

// (0) project, (1) name, (2) sheet music, (3) click track, (4) score order

func partsToCSV(src []parts.Part) []byte {
	records := make([][]string, len(src))
	for i := range records {
		records[i] = make([]string, 5)
		records[i][0] = src[i].ID.Project
		records[i][1] = src[i].ID.Name
		records[i][2] = src[i].Sheets.Key()
		records[i][3] = src[i].Clix.Key()
		records[i][4] = strconv.Itoa(src[i].Meta.ScoreOrder)
	}

	var buf bytes.Buffer
	if err := csv.NewWriter(&buf).WriteAll(records); err != nil {
		logger.WithError(err).Error("csvWriter.WriteAll() failed")
	}
	return buf.Bytes()
}

func partsFromCSV(src io.Reader) []parts.Part {
	csvReader := csv.NewReader(src)
	var dest []parts.Part

	records, err := csvReader.ReadAll()
	if err != nil {
		logger.WithError(err).Error("csvReader.ReadAll() failed")
	}
	for _, record := range records {
		id := parts.ID{
			Project: record[0],
			Name:    record[1],
		}
		sheets := make([]parts.Link, 0, 1)
		if record[2] != "" {
			sheets = append(sheets, parts.Link{ObjectKey: record[2], CreatedAt: time.Now()})
		}
		clix := make([]parts.Link, 0, 1)
		if record[3] != "" {
			clix = append(clix, parts.Link{ObjectKey: record[3], CreatedAt: time.Now()})
		}
		scoreOrder, _ := strconv.Atoi(record[4])
		dest = append(dest, parts.Part{
			ID:     id,
			Sheets: sheets,
			Clix:   clix,
			Meta: parts.Meta{
				ScoreOrder: scoreOrder,
			},
		})
	}
	return dest
}
