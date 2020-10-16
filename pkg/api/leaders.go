package api

import (
	"bytes"
	"context"
	"net/http"
)

type LeadersView struct {
	SpreadSheetID string
}

func (x LeadersView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	leaders, err := listLeaders(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	renderLeadersView(w, ctx, leaders)
}

func renderLeadersView(w http.ResponseWriter, ctx context.Context, leaders []Leader) {
	opts := NewNavBarOpts(ctx)
	page := struct {
		NavBar   NavBarOpts
		Leaders []Leader
	}{
		NavBar:   opts,
		Leaders: leaders,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &page, PublicFiles+"/leaders.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
