package slash_command

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/when2meet"
)

func When2meetCommandOptions(context.Context) ([]discord.ApplicationCommandOption, error) {
	return []discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionTypeString,
			Name:        "event_name",
			Description: "A name for the event.",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionTypeString,
			Name:        "start_date",
			Description: "Start Date (ex 2021-02-04)",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionTypeString,
			Name:        "end_date",
			Description: "End Date (ex 2021-02-05)",
			Required:    true,
		},
	}, nil
}

func when2meetInteractionHandler(_ context.Context, interaction discord.Interaction) discord.InteractionResponse {
	var eventName, startDate, endDate string
	for _, option := range interaction.Data.Options {
		switch option.Name {
		case "start_date":
			startDate = option.Value
		case "end_date":
			endDate = option.Value
		case "event_name":
			eventName = option.Value
		}
	}
	if eventName == "" || startDate == "" || endDate == "" {
		return InteractionResponseOof
	}

	url, err := when2meet.CreateEvent(eventName, startDate, endDate)
	if err != nil {
		logger.WithError(err).Error("when2meet.CreateEvent() failed", err)
		return InteractionResponseOof
	}
	return InteractionResponseMessage(
		fmt.Sprintf("<@%s> created a [when2meet](%s).", interaction.Member.User.ID, url), true)
}
