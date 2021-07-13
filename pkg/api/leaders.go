package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func LeadersApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &leaders)
}

func AboutMeApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entries, err := readAboutMeEntries(ctx)
	if err != nil {
		logger.WithError(err).Error("readAboutMeEntries() failed")
		internalServerError(w)
		return
	}

	var showEntries []AboutMeEntry
	for _, entry := range entries {
		if entry.Show {
			entry.DiscordID = ""
			showEntries = append(showEntries, entry)
		}
	}

	showEntries = []AboutMeEntry{
		{
			Name:  "Brandon",
			Title: "Executive Director",
			Blurb: "Player of assorted brass, singer of high notes, lead of the Video Team and Communications Team, stressed grad student, and the one who came up with this crazy idea.",
			Show:  true,
		},
		{
			Name:  "Jackson",
			Blurb: "I like to make music and code things.",
			Show:  true,
		},
		{
			Name:  "Jacob",
			Blurb: "Making music, having fun, and making more music.",
			Show:  true,
		},
		{
			Name:  "Jerome",
			Blurb: "I look good.",
			Show:  true,
		},
		{
			Name: "Jose",
			Show: true,
		},
		{
			Name:  "Joselyn",
			Blurb: "Coffee.",
			Show:  true,
		},
	}

	jsonEncode(w, showEntries)
}
