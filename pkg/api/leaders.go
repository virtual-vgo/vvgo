package api

import (
	"encoding/json"
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

	switch r.Method {
	case http.MethodGet:
		jsonEncode(w, &leaders)

	case http.MethodPost:
		var leader sheets.Leader
		if err := json.NewDecoder(r.Body).Decode(&leader); err != nil {
			logger.WithError(err).Error("json.Decode() failed")
			badRequest(w, "invalid json")
			return
		}
		if leader.DiscordID == "" {
			badRequest(w, "discord_id required")
			return
		}

		isNew := true
		for i := range leaders {
			if leaders[i].DiscordID == leader.DiscordID {
				isNew = false
				leaders[i] = leader
				break
			}
		}
		if isNew {
			leaders = append(leaders, leader)
		}

		identity := IdentityFromContext(ctx)
		if identity.DiscordID != leader.DiscordID {
			logger.Infof("wanted `%s`, got `%s`", identity.DiscordID, leader.DiscordID)
			unauthorized(w)
			return
		}

		if err := sheets.WriteLeaders(ctx, leaders); err != nil {
			logger.WithError(err).Error("sheets.WriteLeaders() failed")
			internalServerError(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":"true"}`))
	}
}
