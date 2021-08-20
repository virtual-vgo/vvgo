package slash_command

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	discord2 "github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/server/api/aboutme"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command/foaas"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command/when2meet"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"io"
	"net/http"
	"time"
)

var logger = log.New()

var SlashCommands = []SlashCommand{
	{
		Name:        "beep",
		Description: "Send a beep.",
		Handler:     beepInteractionHandler,
	},
	{
		Name:        "parts",
		Description: "Parts link for a project.",
		Options:     partsCommandOptions,
		Handler:     partsInteractionHandler,
	},
	{
		Name:        "submit",
		Description: "Submission link for a project.",
		Options:     submitCommandOptions,
		Handler:     submitInteractionHandler,
	},
	{
		Name:        "fuckoff",
		Description: "A modern solution to the common problem of telling people to fuck off.",
		Handler:     fuckoffInteractionHandler,
	},
	{
		Name:        "when2meet",
		Description: "Make a when2meet link.",
		Options:     when2meetCommandOptions,
		Handler:     when2meetInteractionHandler,
	},
	{
		Name:        "aboutme",
		Description: "Manage your about me blurb on the vvgo website.",
		Options:     aboutmeCommandOptions,
		Handler:     aboutmeInteractionHandler,
	},
}

var InteractionResponseOof = interactionResponseMessage("oof please try again 😅", true)
var InteractionResponseGalaxyBrain = interactionResponseMessage("this interaction is too galaxy brain for me 😥", true)

func Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	for _, command := range SlashCommands {
		<-timer.C
		if err := command.Create(ctx); err != nil {
			logger.MethodFailure(ctx, "SlashCommand.Create", err)
			helpers.InternalServerError(w)
			return
		} else {
			logger.Info(command.Name, "command created")
		}
	}
	http.Redirect(w, r, "/slash_commands", http.StatusFound)
}

func List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commands, err := discord2.GetApplicationCommands(ctx)
	if err != nil {
		logger.WithError(err).Error("discord.GetApplicationCommands() failed")
		helpers.InternalServerError(w)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body bytes.Buffer
	_, _ = body.ReadFrom(r.Body)

	publicKey, _ := hex.DecodeString(discord2.ClientPublicKey)
	if len(publicKey) == 0 {
		logger.Error("invalid discord public key")
		helpers.InternalServerError(w)
		return
	}

	signature, _ := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if len(signature) == 0 {
		helpers.BadRequest(w, "invalid signature")
		return
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if len(timestamp) == 0 {
		helpers.BadRequest(w, "invalid signature timestamp")
		return
	}

	if ed25519.Verify(publicKey, []byte(timestamp+body.String()), signature) == false {
		helpers.Unauthorized(w)
		return
	}

	var interaction discord2.Interaction
	if err := json.NewDecoder(&body).Decode(&interaction); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		helpers.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	response, ok := HandleInteraction(ctx, interaction)
	if !ok {
		helpers.BadRequest(w, "unsupported interaction type")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		helpers.InternalServerError(w)
	}
}

func HandleInteraction(ctx context.Context, interaction discord2.Interaction) (discord2.InteractionResponse, bool) {
	switch interaction.Type {
	case discord2.InteractionTypePing:
		return discord2.InteractionResponse{Type: discord2.InteractionCallbackTypePong}, true
	case discord2.InteractionTypeApplicationCommand:
		for _, command := range SlashCommands {
			if interaction.Data.Name == command.Name {
				return command.Handler(ctx, interaction), true
			}
		}
		return InteractionResponseGalaxyBrain, true
	default:
		return discord2.InteractionResponse{}, false
	}
}

type SlashCommand struct {
	Name        string
	Description string
	Options func(context.Context) ([]discord2.ApplicationCommandOption, error)
	Handler InteractionHandler
}

type InteractionHandler func(context.Context, discord2.Interaction) discord2.InteractionResponse

func (x SlashCommand) Create(ctx context.Context) (err error) {
	var options []discord2.ApplicationCommandOption
	if x.Options != nil {
		options, err = x.Options(ctx)
		if err != nil {
			return err
		}
	}
	params := discord2.CreateApplicationCommandParams{
		Name:        x.Name,
		Description: x.Description,
		Options:     options,
	}
	_, err = discord2.CreateApplicationCommand(ctx, params)
	return err
}

func beepInteractionHandler(context.Context, discord2.Interaction) discord2.InteractionResponse {
	return interactionResponseMessage("boop", false)
}

func partsCommandOptions(ctx context.Context) ([]discord2.ApplicationCommandOption, error) {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord2.ApplicationCommandOption{projectCommandOption(projects.Current())}, nil
}

func partsInteractionHandler(ctx context.Context, interaction discord2.Interaction) discord2.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		logger.WithError(err).Error("sheets.ListProjects() failed")
		return InteractionResponseOof
	}

	project, ok := projects.Get(projectName)
	if !ok {
		return InteractionResponseOof
	}

	description := fmt.Sprintf(`· Parts are [here!](https://vvgo.org%s)
· Submit files [here!](%s)
· Submission Deadline: %s.`,
		project.PartsPage(), project.SubmissionLink, project.SubmissionDeadline)

	embed := discord2.Embed{
		Title:       project.Title,
		Type:        discord2.EmbedTypeRich,
		Description: description,
		Url:         "https://vvgo.org" + project.PartsPage(),
		Color:       0x8C17D9,
		Footer:      &discord2.EmbedFooter{Text: "Bottom text."},
	}
	return discord2.InteractionResponse{
		Type: discord2.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord2.InteractionApplicationCommandCallbackData{
			Embeds: []discord2.Embed{embed},
		},
	}
}

func submitCommandOptions(ctx context.Context) ([]discord2.ApplicationCommandOption, error) {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord2.ApplicationCommandOption{projectCommandOption(projects.Current())}, nil
}

func projectCommandOption(projects sheets.Projects) discord2.ApplicationCommandOption {
	var choices []discord2.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord2.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return discord2.ApplicationCommandOption{
		Type:        discord2.ApplicationCommandOptionTypeString,
		Name:        "project",
		Description: "Name of the project",
		Required:    true,
		Choices:     choices,
	}
}

func submitInteractionHandler(ctx context.Context, interaction discord2.Interaction) discord2.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	var content string
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		logger.WithError(err).Error("sheets.ListProjects() failed")
	} else if project, ok := projects.Get(projectName); ok {
		content = fmt.Sprintf(`[Submit here](%s) for %s. Submission Deadline is `, project.SubmissionLink, project.Title)
	}

	if content == "" {
		return InteractionResponseOof
	}
	return interactionResponseMessage(content, true)
}

func fuckoffInteractionHandler(_ context.Context, interaction discord2.Interaction) discord2.InteractionResponse {
	content, _ := foaas.FuckOff(fmt.Sprintf("<@%s>", interaction.Member.User.ID))
	if content == "" {
		return InteractionResponseOof
	}
	return interactionResponseMessage(content, true)
}

func when2meetCommandOptions(context.Context) ([]discord2.ApplicationCommandOption, error) {
	return []discord2.ApplicationCommandOption{
		{
			Type:        discord2.ApplicationCommandOptionTypeString,
			Name:        "event_name",
			Description: "A name for the event.",
			Required:    true,
		},
		{
			Type:        discord2.ApplicationCommandOptionTypeString,
			Name:        "start_date",
			Description: "Start Date (ex 2021-02-04)",
			Required:    true,
		},
		{
			Type:        discord2.ApplicationCommandOptionTypeString,
			Name:        "end_date",
			Description: "End Date (ex 2021-02-05)",
			Required:    true,
		},
	}, nil
}

func when2meetInteractionHandler(_ context.Context, interaction discord2.Interaction) discord2.InteractionResponse {
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
	return interactionResponseMessage(
		fmt.Sprintf("<@%s> created a [when2meet](%s).", interaction.Member.User.ID, url), true)
}

func aboutmeCommandOptions(context.Context) ([]discord2.ApplicationCommandOption, error) {
	return []discord2.ApplicationCommandOption{
		{
			Type:        discord2.ApplicationCommandOptionTypeSubCommand,
			Name:        "summary",
			Description: "Get a summary of your aboutme information on the vvgo website.",
		},
		{
			Type:        discord2.ApplicationCommandOptionTypeSubCommand,
			Name:        "show",
			Description: "Show your aboutme information on the vvgo website.",
		},
		{
			Type:        discord2.ApplicationCommandOptionTypeSubCommand,
			Name:        "hide",
			Description: "Hide your aboutme information from the vvgo website.",
		},
		{
			Type:        discord2.ApplicationCommandOptionTypeSubCommand,
			Name:        "update",
			Description: "Update your aboutme information.",
			Options: []discord2.ApplicationCommandOption{
				{
					Type:        discord2.ApplicationCommandOptionTypeString,
					Name:        "name",
					Description: "Your name.",
				},
				{
					Type:        discord2.ApplicationCommandOptionTypeString,
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

	if hasRole(discord2.VVGOExecutiveDirectorRoleID) {
		return "Executive Director"
	} else if hasRole(discord2.VVGOProductionDirectorRoleID) {
		return "Production Director"
	} else if hasRole(discord2.VVGOProductionTeamRoleID) {
		return "Production Team"
	} else {
		return ""
	}
}

func aboutmeInteractionHandler(ctx context.Context, interaction discord2.Interaction) discord2.InteractionResponse {
	userId := interaction.Member.User.ID.String()
	title := getAboutMeTitleFromRoles(interaction.Member.Roles)

	isProduction := false
	for _, role := range interaction.Member.Roles {
		if role == discord2.VVGOProductionTeamRoleID {
			isProduction = true
		}
	}

	if !isProduction {
		return interactionResponseMessage("Sorry, this tool is only for production teams. :bow:", true)
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

func summaryAboutMe(entries map[string]aboutme.Entry, userId string) discord2.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		message := fmt.Sprintf("**%s** ~ %s ~\n", entry.Name, entry.Blurb)
		message += "Use `/aboutme update` to make changes.\n"
		if entry.Show {
			message += "Your name and blurb are visible on https://vvgo.org/about. Use `/aboutme hide` to hide it."
		} else {
			message += "Your name and blurb are not visible on https://vvgo.org/about."
		}
		return interactionResponseMessage(message, true)
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func hideAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string) discord2.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = false
		entries[userId] = entry
		if err := aboutme.WriteEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are hidden from https://vvgo.org/about.", true)
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func showAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string) discord2.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = true
		entries[userId] = entry
		if err := aboutme.WriteEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are visible on https://vvgo.org/about.", true)
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func updateAboutme(ctx context.Context, entries map[string]aboutme.Entry, userId string, title string, option discord2.ApplicationCommandInteractionDataOption) discord2.InteractionResponse {
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
	return interactionResponseMessage(":person_gesturing_ok: It is written.", true)
}

func interactionResponseMessage(text string, ephemeral bool) discord2.InteractionResponse {
	var flags int
	if ephemeral {
		flags = discord2.InteractionApplicationCommandCallbackDataFlagEphemeral
	}
	return discord2.InteractionResponse{
		Type: discord2.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord2.InteractionApplicationCommandCallbackData{
			Content: text,
			Flags:   flags,
		},
	}
}
