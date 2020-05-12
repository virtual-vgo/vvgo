package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"strconv"
)

type BackupDocument struct {
	Parts []parts.Part
}

type BackupHandler struct {
	Storage
}

func (x BackupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_handler")
	defer span.Send()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	gotParts, err := x.Parts.List(ctx)
	if err != nil {
		logger.WithError(err).Error("Parts.List() failed")
		internalServerError(w)
		return
	}

	jsonEncode(w, &BackupDocument{Parts: gotParts})
}

type RestoreHandler struct {
	Storage
}

func (x RestoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_handler")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	// check if we should truncate existing dbs
	if truncateParts, _ := strconv.ParseBool(r.FormValue("truncate_parts")); truncateParts {
		if err := x.Parts.DeleteAll(ctx); err != nil {
			logger.WithError(err).Error("Parts.DeleteAll() failed")
			internalServerError(w)
		}
	}

	// parse the document
	var document BackupDocument
	if err := json.NewDecoder(r.Body).Decode(&document); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		badRequest(w, err.Error())
	}

	// save the new parts
	if err := x.Parts.Save(ctx, document.Parts); err != nil {
		logger.WithError(err).Error("Parts.Save() failed")
		internalServerError(w)
	}
}
