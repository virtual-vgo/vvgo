package api

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"sort"
	"strings"
	"time"
)

const SkywardSwordStatsChannelID = "700792848253059142" // "844859046863044638"

func SkywardSwordIntentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// set a rate-limit on this handler
	var lockStatus string
	err := redis.Do(ctx, redis.Cmd(&lockStatus, "SET", "intent_stats:skyward_sword:lock", "locked", "NX", "EX", "5"))
	if err != nil {
		logger.WithError(err).Error("redis.Do() failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(lockStatus)
	if lockStatus != "OK" {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	err = updateIntentMessage(ctx)
	if err != nil {
		logger.WithError(err).Error("discordClient.CreateMessage() failed")
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

	// generate the message content
	var content = "*Here's what parts people say they intend to play!*\n"
	for _, part := range parts { // dont loop over map to preserve sorting
		names := intentMap[part]
		sort.Strings(names)
		content += fmt.Sprintf("> **%s (%d):** %s\n", part, len(names), strings.Join(names, ", "))
	}
	content += "\nClick here to update: https://vvgo.org/api/v1/update_stats" +
		"\n_Last Updated: " + time.Now().Format(time.UnixDate) + "_"

	// batch the content into separate messages
	discordClient := discord.NewClient(ctx)
	lines := strings.Split(content, "\n")
	var nextContent string
	var messageIds []string
	for _, line := range lines {
		if len(nextContent)+len(line) > 1500 {
			message, err := discordClient.CreateMessage(ctx, SkywardSwordStatsChannelID,
				discord.CreateMessageParams{Content: nextContent})
			if err != nil {
				return err
			}
			messageIds = append(messageIds, message.Id)
			nextContent = ""
		}
		nextContent += line + "\n"
	}
	message, err := discordClient.CreateMessage(ctx, SkywardSwordStatsChannelID,
		discord.CreateMessageParams{Content: nextContent})
	if err != nil {
		return err
	}
	messageIds = append(messageIds, message.Id)

	// post the final message
	finalContent := `
**Intent Form:** https://docs.google.com/forms/d/e/1FAIpQLSchQa04TaiVWWvYGYAkCfCFqMvrxBy-h2DN1IjdoQ9qpRtuAQ/viewform
Please use this form to indicate what part you intend to record. The above post will be kept as up-to-date as Section Leader Chicken receives your responses.
`
	message, err = discordClient.CreateMessage(ctx, SkywardSwordStatsChannelID,
		discord.CreateMessageParams{Content: finalContent})
	if err != nil {
		return err
	}
	messageIds = append(messageIds, message.Id)

	// delete the old messages
	var oldMessageIdsRaw string
	if err := redis.Do(ctx, redis.Cmd(&oldMessageIdsRaw, "GET", "intent_stats:skyward_sword:message_ids")); err != nil {
		logger.WithError(err).Error("redis.Do() failed")
	}
	if len(oldMessageIdsRaw) > 0 {
		err := discordClient.BulkDeleteMessages(ctx, SkywardSwordStatsChannelID,
			discord.BulkDeleteMessagesParams{Messages: strings.Split(oldMessageIdsRaw, ",")})
		if err != nil {
			logger.WithError(err).Info("discordClient.BulkDeleteMessages failed")
		}
	}

	// store the new ids in redis
	if err := redis.Do(ctx, redis.Cmd(nil, "SET", "intent_stats:skyward_sword:message_ids", strings.Join(messageIds, ","))); err != nil {
		logger.WithError(err).Error("redis.Do() failed")
	}
	return nil
}
