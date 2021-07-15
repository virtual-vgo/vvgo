package api

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/error_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/foaas"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/when2meet"
	"io"
	"net/http"
	"strings"
)

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

func CreateSlashCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	for _, command := range SlashCommands {
		handleError(command.Create(ctx)).
			logError("SlashCommand.Create() failed").
			logSuccess(command.Name + " command created")
	}
	http.Redirect(w, r, "/slash_commands", http.StatusFound)
}

func ViewSlashCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commands, err := discord.NewClient(ctx).GetApplicationCommands(ctx)
	if err != nil {
		logger.WithError(err).Error("discord.GetApplicationCommands() failed")
		internalServerError(w)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

func HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body bytes.Buffer
	_, _ = body.ReadFrom(r.Body)

	publicKey, _ := hex.DecodeString(discord.ClientPublicKey)
	if len(publicKey) == 0 {
		logger.Error("invalid discord public key")
		internalServerError(w)
		return
	}

	signature, _ := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if len(signature) == 0 {
		badRequest(w, "invalid signature")
		return
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if len(timestamp) == 0 {
		badRequest(w, "invalid signature timestamp")
		return
	}

	if ed25519.Verify(publicKey, []byte(timestamp+body.String()), signature) == false {
		unauthorized(w)
		return
	}

	var interaction discord.Interaction
	if err := json.NewDecoder(&body).Decode(&interaction); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	response, ok := HandleInteraction(ctx, interaction)
	if !ok {
		badRequest(w, "unsupported interaction type")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		internalServerError(w)
	}
}

func HandleInteraction(ctx context.Context, interaction discord.Interaction) (discord.InteractionResponse, bool) {
	switch interaction.Type {
	case discord.InteractionTypePing:
		return discord.InteractionResponse{Type: discord.InteractionCallbackTypePong}, true
	case discord.InteractionTypeApplicationCommand:
		for _, command := range SlashCommands {
			if interaction.Data.Name == command.Name {
				return command.Handler(ctx, interaction), true
			}
		}
		return InteractionResponseGalaxyBrain, true
	default:
		return discord.InteractionResponse{}, false
	}
}

type SlashCommand struct {
	Name        string
	Description string
	Options     func(context.Context) ([]discord.ApplicationCommandOption, error)
	Handler     InteractionHandler
}

type InteractionHandler func(context.Context, discord.Interaction) discord.InteractionResponse

func (x SlashCommand) Create(ctx context.Context) (err error) {
	var options []discord.ApplicationCommandOption
	if x.Options != nil {
		options, err = x.Options(ctx)
		if err != nil {
			return err
		}
	}
	params := discord.CreateApplicationCommandParams{
		Name:        x.Name,
		Description: x.Description,
		Options:     options,
	}
	_, err = discord.NewClient(ctx).CreateApplicationCommand(ctx, params)
	return err
}

func beepInteractionHandler(context.Context, discord.Interaction) discord.InteractionResponse {
	return interactionResponseMessage("boop", false)
}

func partsCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord.ApplicationCommandOption{projectCommandOption(projects.Current())}, nil
}

func partsInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
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

	embed := discord.Embed{
		Title:       project.Title,
		Type:        discord.EmbedTypeRich,
		Description: description,
		Url:         "https://vvgo.org" + project.PartsPage(),
		Color:       0x8C17D9,
		Footer:      &discord.EmbedFooter{Text: "Bottom text."},
	}
	return discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Embeds: []discord.Embed{embed},
		},
	}
}

func submitCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord.ApplicationCommandOption{projectCommandOption(projects.Current())}, nil
}

func projectCommandOption(projects sheets.Projects) discord.ApplicationCommandOption {
	var choices []discord.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return discord.ApplicationCommandOption{
		Type:        discord.ApplicationCommandOptionTypeString,
		Name:        "project",
		Description: "Name of the project",
		Required:    true,
		Choices:     choices,
	}
}

func submitInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
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

func fuckoffInteractionHandler(_ context.Context, interaction discord.Interaction) discord.InteractionResponse {
	content, _ := foaas.FuckOff(fmt.Sprintf("<@%s>", interaction.Member.User.ID))
	if content == "" {
		return InteractionResponseOof
	}
	return interactionResponseMessage(content, true)
}

func when2meetCommandOptions(context.Context) ([]discord.ApplicationCommandOption, error) {
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
	return interactionResponseMessage(
		fmt.Sprintf("<@%s> created a [when2meet](%s).", interaction.Member.User.ID, url), true)
}

type AboutMeEntry struct {
	DiscordID string `json:"discord_id,omitempty"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Blurb     string `json:"blurb"`
	Show      bool   `json:"show"`
}

func readAboutMeEntries(ctx context.Context, keys []string) (map[string]AboutMeEntry, error) {
	if keys == nil {
		buf := make(map[string]string)
		cmd := "HGETALL"
		args := []string{"about_me:entries"}
		if err := redis.Do(ctx, redis.Cmd(&buf, cmd, args...)); err != nil {
			return nil, error_wrappers.RedisDoFailed(err)
		}
		dest := make(map[string]AboutMeEntry)
		for _, entryJson := range buf {
			var entry AboutMeEntry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, error_wrappers.JsonDecodeFailed(err)
			}
			dest[entry.DiscordID] = entry
		}
		return dest, nil
	} else {
		var buf []string
		cmd := "HMGET"
		args := append([]string{"about_me:entries"}, keys...)
		if err := redis.Do(ctx, redis.Cmd(&buf, cmd, args...)); err != nil {
			return nil, error_wrappers.RedisDoFailed(err)
		}
		dest := make(map[string]AboutMeEntry)
		for _, entryJson := range buf {
			var entry AboutMeEntry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, error_wrappers.JsonDecodeFailed(err)
			}
			dest[entry.DiscordID] = entry
		}
		return dest, nil
	}
}

func writeAboutMeEntries(ctx context.Context, src map[string]AboutMeEntry) error {
	if len(src) == 0 {
		logger.Warnf("writeAboutMeEntries: there's nothing to write!")
		return nil
	}

	args := []string{"about_me:entries"}
	for id, entry := range src {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(entry); err != nil {
			return error_wrappers.JsonEncodeFailed(err)
		}
		args = append(args, id, buf.String())
	}

	if err := redis.Do(ctx, redis.Cmd(nil, "HMSET", args...)); err != nil {
		return error_wrappers.RedisDoFailed(err)
	}
	return nil
}

func deleteAboutmeEntries(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	args := append([]string{"about_me:entries"}, keys...)
	if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
		return error_wrappers.RedisDoFailed(err)
	}
	return nil
}

func aboutmeCommandOptions(context.Context) ([]discord.ApplicationCommandOption, error) {
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

func aboutmeInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	userId := interaction.Member.User.ID.String()
	title := getAboutMeTitleFromRoles(interaction.Member.Roles)

	isProduction := false
	for _, role := range interaction.Member.Roles {
		if role == discord.VVGOProductionTeamRoleID {
			isProduction = true
		}
	}

	if !isProduction {
		return interactionResponseMessage("Sorry, this tool is only for production teams. :bow:", true)
	}

	entries, err := readAboutMeEntries(ctx, []string{userId})
	if err != nil && !errors.Is(err, io.EOF) {
		logger.WithError(err).Error("readAboutMeEntries() failed")
		return InteractionResponseOof
	}

	if entries == nil {
		entries = make(map[string]AboutMeEntry)
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

func summaryAboutMe(entries map[string]AboutMeEntry, userId string) discord.InteractionResponse {
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

func hideAboutme(ctx context.Context, entries map[string]AboutMeEntry, userId string) discord.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = false
		entries[userId] = entry
		if err := writeAboutMeEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are hidden from https://vvgo.org/about.", true)
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func showAboutme(ctx context.Context, entries map[string]AboutMeEntry, userId string) discord.InteractionResponse {
	if entry, ok := entries[userId]; ok {
		entry.Show = true
		entries[userId] = entry
		if err := writeAboutMeEntries(ctx, entries); err != nil {
			logger.WithError(err).Error("writeAboutMeEntries() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are visible on https://vvgo.org/about.", true)
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:", true)
}

func updateAboutme(ctx context.Context, entries map[string]AboutMeEntry, userId string, title string, option discord.ApplicationCommandInteractionDataOption) discord.InteractionResponse {
	updateEntry := func(entry AboutMeEntry) AboutMeEntry {
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
		entries[userId] = updateEntry(AboutMeEntry{DiscordID: userId})
	}

	if err := writeAboutMeEntries(ctx, entries); err != nil {
		logger.WithError(err).Error("writeAboutMeEntries() failed")
		return InteractionResponseOof
	}
	return interactionResponseMessage(":person_gesturing_ok: It is written.", true)
}

func interactionResponseMessage(text string, ephemeral bool) discord.InteractionResponse {
	var flags int
	if ephemeral {
		flags = discord.InteractionApplicationCommandCallbackDataFlagEphemeral
	}
	return discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: text,
			Flags:   flags,
		},
	}
}
