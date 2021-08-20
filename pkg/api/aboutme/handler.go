package aboutme

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
)

var logger = log.New()

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		entries, err := ReadEntries(ctx, nil)
		if err != nil {
			logger.WithError(err).Error("ReadEntries() failed")
			helpers.InternalServerError(w)
			return
		}
		var showEntries []Entry
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

		var newEntries []Entry
		if err := json.NewDecoder(r.Body).Decode(&newEntries); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			helpers.BadRequest(w, "invalid json")
			return
		}
		if len(newEntries) == 0 {
			return
		}

		entriesMap := make(map[string]Entry)
		for _, entry := range newEntries {
			if entry.DiscordID != "" {
				entriesMap[entry.DiscordID] = entry
			}
		}

		if err := WriteEntries(ctx, entriesMap); err != nil {
			logger.WithError(err).Error("WriteEntries() failed")
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

		if err := DeleteEntries(ctx, keys); err != nil {
			logger.WithError(err).Error("deleteAboutMeEntries() failed")
			helpers.InternalServerError(w)
			return
		}
	}
}
