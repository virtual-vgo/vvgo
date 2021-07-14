package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/login"
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

	writeEntriesToResponse := func(entries map[string]AboutMeEntry) {
		var showEntries []AboutMeEntry
		isLeader := IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader)
		for _, entry := range entries {
			if entry.Show {
				if isLeader == false {
					entry.DiscordID = ""
				}
				showEntries = append(showEntries, entry)
			}
		}
		jsonEncode(w, showEntries)
	}

	switch r.Method {
	case http.MethodGet:
		entries, err := readAboutMeEntries(ctx)
		if err != nil {
			logger.WithError(err).Error("readAboutMeEntries() failed")
			internalServerError(w)
			return
		}
		writeEntriesToResponse(entries)

	case http.MethodPost:
		if IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader) == false {
			unauthorized(w)
			return
		}

		var newEntries []AboutMeEntry
		if err := json.NewDecoder(r.Body).Decode(&newEntries); err != nil {
			logJsonDecodeErr(err)
			badRequest(w, "invalid json")
			return
		}
		if len(newEntries) == 0 {
			return
		}

		entriesMap, err := readAboutMeEntries(ctx)
		if err != nil {
			logger.WithError(err).Error("readAboutMeEntries() failed")
			internalServerError(w)
			return
		}

		for _, entry := range newEntries {
			if entry.DiscordID != "" {
				entriesMap[entry.DiscordID] = entry
			}
		}

		if err := writeAboutMeEntries(ctx, entriesMap); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			internalServerError(w)
			return
		}
		writeEntriesToResponse(entriesMap)

	case http.MethodDelete:
		if IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader) == false {
			unauthorized(w)
			return
		}

		var delEntries []AboutMeEntry
		if err := json.NewDecoder(r.Body).Decode(&delEntries); err != nil {
			logJsonDecodeErr(err)
			badRequest(w, "invalid json")
			return
		}
		if len(delEntries) == 0 {
			return
		}

		entriesMap, err := readAboutMeEntries(ctx)
		if err != nil {
			logger.WithError(err).Error("readAboutMeEntries() failed")
			internalServerError(w)
			return
		}

		for _, entry := range delEntries {
			if entry.DiscordID != "" {
				delete(entriesMap, entry.DiscordID)
			}
		}

		if err := writeAboutMeEntries(ctx, entriesMap); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			internalServerError(w)
			return
		}
		writeEntriesToResponse(entriesMap)
	}
}

func logJsonDecodeErr(err error) {
	logger.WithError(err).Error("json.Decode() failed")
}
