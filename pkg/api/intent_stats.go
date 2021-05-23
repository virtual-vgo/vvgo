package api

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"sort"
	"strings"
	"time"
)

const SkywardSwordIntentMessageID = "846092722481397771"
const SkywardSwordStatsChannelID = "844859046863044638"

func SkywardSwordIntentHandler(w http.ResponseWriter, r *http.Request) {
	err := updateIntentMessage(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateIntentMessage(ctx context.Context) error {
	intents, err := sheets.ListSkywardSwordIntents(ctx)
	if err != nil {
		return err
	}

	// collapse into nicer format
	intentMap := make(map[string][]string)
	for _, intent := range intents {
		part := intent.IntendToRecord
		intentMap[part] = append(intentMap[part], strings.TrimSpace(intent.CreditedName))
	}

	var parts []string
	for k := range intentMap {
		parts = append(parts, k)
	}
	sort.Strings(parts)

	var message = "*Here's what parts people say they intend to play!*\n"
	for _, part := range parts { // dont loop over map to preserve sorting
		names := intentMap[part]
		sort.Strings(names)
		message += fmt.Sprintf("> **%s (%d):** %s\n", part, len(names), strings.Join(names, ", "))
	}
	message += "\n_Last Updated: " + time.Now().Format(time.UnixDate) + "_"
	params := discord.EditMessageParams{Content: message}

	return discord.
		NewClient(ctx).
		EditMessage(ctx, SkywardSwordStatsChannelID, SkywardSwordIntentMessageID, params)
}
