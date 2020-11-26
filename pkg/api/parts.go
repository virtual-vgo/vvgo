package api

import (
	"bytes"
	"net/http"
)

type PartView struct {
	SpreadsheetID string
	*Database
}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, nil, "parts.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
