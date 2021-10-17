package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/models/aboutme"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Aboutme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		entries, err := aboutme.ReadEntries(ctx, nil)
		if err != nil {
			logger.MethodFailure(ctx, "aboutme.ReadEntries", err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
		var showEntries []aboutme.Entry
		isLeader := login.IdentityFromContext(ctx).HasRole(models.RoleVVGOLeader)
		for _, entry := range entries {
			if entry.Show {
				if isLeader == false {
					entry.DiscordID = ""
				}
				showEntries = append(showEntries, entry)
			}
		}
		http_helpers.JsonEncode(w, showEntries)

	case http.MethodPost:
		if login.IdentityFromContext(ctx).HasRole(models.RoleVVGOLeader) == false {
			http_helpers.Unauthorized(ctx, w)
			return
		}

		var newEntries []aboutme.Entry
		if err := json.NewDecoder(r.Body).Decode(&newEntries); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			http_helpers.BadRequest(ctx, w, "invalid json")
			return
		}
		if len(newEntries) == 0 {
			return
		}

		entriesMap := make(map[string]aboutme.Entry)
		for _, entry := range newEntries {
			if entry.DiscordID != "" {
				entriesMap[entry.DiscordID] = entry
			}
		}

		if err := aboutme.WriteEntries(ctx, entriesMap); err != nil {
			logger.MethodFailure(ctx, "aboutme.WriteEntries", err)
			http_helpers.InternalServerError(ctx, w)
			return
		}

	case http.MethodDelete:
		if login.IdentityFromContext(ctx).HasRole(models.RoleVVGOLeader) == false {
			http_helpers.Unauthorized(ctx, w)
			return
		}

		var keys []string
		if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			http_helpers.BadRequest(ctx, w, "invalid json")
			return
		}
		if len(keys) == 0 {
			return
		}

		if err := aboutme.DeleteEntries(ctx, keys); err != nil {
			logger.MethodFailure(ctx, "aboutme.DeleteEntries", err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
	}
}
