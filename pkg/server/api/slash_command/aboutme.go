package slash_command

import (
	"context"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models/aboutme"
	"io"
)

func AboutmeCommandOptions(context.Context) ([]discord.ApplicationCommandOption, error) {
	return []discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "summary",
			Description: "Get a summary of your aboutme information on the vvgo website.",
		},
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "show",
			Description: "Show your aboutme information on the vvgo website.",
		},
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "hide",
			Description: "Hide your aboutme information from the vvgo website.",
		},
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "update",
			Description: "Update your aboutme information.",
			Options: []discord.ApplicationCommandOption{
				{
					Type:        discord.ApplicationCommandOptionTypeString,
					Name:        "name",
					Description: "Your name.",
				},
				{
					Type:        discord.ApplicationCommandOptionTypeString,
					Name:        "blurb",
					Description: "A blurb about yourself.",
				},
			},
		},
	}, nil
}

func getAboutMeTitleFromRoles(roles []string) string {
	hasRole := func(want string) bool {
		for _, role := range roles {
			if role == want {
				return true
			}
		}
		return false
	}

	if hasRole(discord.VVGOExecutiveDirectorRoleID) {
		return "Executive Director"
	} else if hasRole(discord.VVGOProductionDirectorRoleID) {
		return "Production Director"
	} else if hasRole(discord.VVGOProductionTeamRoleID) {
		return "Production Team"
	} else {
		return ""
	}
}

func AboutmeInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	userId := interaction.Member.User.ID.String()
	title := getAboutMeTitleFromRoles(interaction.Member.Roles)

	isProduction := false
	for _, role := range interaction.Member.Roles {
		if role == discord.VVGOProductionTeamRoleID {
			isProduction = true
		}
	}

	if !isProduction {
		return InteractionResponseMessage("Sorry, this tool is only for production teams. :bow:", true)
	}

	entries, err := aboutme.ReadEntries(ctx, []string{userId})
	if err != nil && !errors.Is(err, io.EOF) {
		logger.WithError(err).Error("readAboutMeEntries() failed")
		return InteractionResponseOof
	}

	if entries == nil {
		entries = make(map[string]aboutme.Entry)
	}

	for _, option := range interaction.Data.Options {
		switch option.Name {
		case "summary":
			return summaryAboutMe(entries, userId)
		case "hide":
			return hideAboutme(ctx, entries, userId)
		case "show":
			return showAboutme(ctx, entries, userId)
		case "update":
			return updateAboutme(ctx, entries, userId, title, option)
		}
	}
	return InteractionResponseOof
}

func summaryAboutMe(entries map[string]aboutme.Entry, userId string) discord.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		message := fmt.Sprintf("**%s** ~ %s ~\n", entry.Name, entry.Blurb)
		message += "Use `/aboutme update` to make changes.\n"
		if entry.Show {
			message += "Your name and blurb are visible on https://vvgo.org/about. Use `/aboutme hide` to hide it."
		} else {
			message += "Your name and blurb are not visible on https://vvgo.org/about."
		}
		return InteractionResponseMessage(message, true)
	}
	return InteractionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func hideAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string) discord.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = false
		entries[userId] = entry
		if err := aboutme.WriteEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return InteractionResponseMessage(":person_gesturing_ok: You are hidden from https://vvgo.org/about.", true)
	}
	return InteractionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func showAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string) discord.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = true
		entries[userId] = entry
		if err := aboutme.WriteEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return InteractionResponseMessage(":person_gesturing_ok: You are visible on https://vvgo.org/about.", true)
	}
	return InteractionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func updateAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string, title string, option discord.ApplicationCommandInteractionDataOption) discord.InteractionResponse {
	updateEntry := func(entry aboutme.Entry) aboutme.Entry {
		entry.Title = title
		for _, option := range option.Options {
			switch option.Name {
			case "name":
				entry.Name = option.Value
			case "blurb":
				entry.Blurb = option.Value
			}
		}
		return entry
	}

	if entry, ok := entries[userId]; ok {
		entries[userId] = updateEntry(entry)
	} else {
		entries[userId] = updateEntry(aboutme.Entry{DiscordID: userId})
	}

	if err := aboutme.WriteEntries(ctx, entries); err != nil {
		logger.WithError(err).Error("writeAboutMeEntries() failed")
		return InteractionResponseOof
	}
	return InteractionResponseMessage(":person_gesturing_ok: It is written.", true)
}
