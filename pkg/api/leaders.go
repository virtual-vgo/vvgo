package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func LeadersApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		helpers.InternalServerError(w)
		return
	}
	helpers.JsonEncode(w, &leaders)
}

func AboutMeApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		entries, err := readAboutMeEntries(ctx, nil)
		if err != nil {
			logger.WithError(err).Error("readAboutMeEntries() failed")
			helpers.InternalServerError(w)
			return
		}
		var showEntries []AboutMeEntry
		isLeader := login.IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader)
		for _, entry := range entries {
			if entry.Show {
				if isLeader == false {
					entry.DiscordID = ""
				}
				showEntries = append(showEntries, entry)
			}
		}
		helpers.JsonEncode(w, showEntries)

	case http.MethodPost:
		if login.IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader) == false {
			helpers.Unauthorized(w)
			return
		}

		var newEntries []AboutMeEntry
		if err := json.NewDecoder(r.Body).Decode(&newEntries); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			helpers.BadRequest(w, "invalid json")
			return
		}
		if len(newEntries) == 0 {
			return
		}

		entriesMap := make(map[string]AboutMeEntry)
		for _, entry := range newEntries {
			if entry.DiscordID != "" {
				entriesMap[entry.DiscordID] = entry
			}
		}

		if err := writeAboutMeEntries(ctx, entriesMap); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			helpers.InternalServerError(w)
			return
		}

	case http.MethodDelete:
		if login.IdentityFromContext(ctx).HasRole(login.RoleVVGOLeader) == false {
			helpers.Unauthorized(w)
			return
		}

		var keys []string
		if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			helpers.BadRequest(w, "invalid json")
			return
		}
		if len(keys) == 0 {
			return
		}

		if err := deleteAboutmeEntries(ctx, keys); err != nil {
			logger.WithError(err).Error("deleteAboutMeEntries() failed")
			helpers.InternalServerError(w)
			return
		}
	}
}
