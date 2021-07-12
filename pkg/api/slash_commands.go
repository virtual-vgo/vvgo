package api

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/foaas"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/when2meet"
	"net/http"
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

func CreateSlashCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//for _, command := range SlashCommands {
	//	handleError(command.Create(ctx)).
	//		logError("SlashCommand.Create() failed").
	//		logSuccess(command.Name + " command created")
	//}
	SlashCommands[5].Create(ctx)
	http.Redirect(w, r, "/slash_commands", http.StatusFound)
}

func ViewSlashCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commands, err := discord.NewClient(ctx).GetApplicationCommands(ctx)
	handleError(err).logError("discord.GetApplicationCommands failed").
		ifError(func(err error) { internalServerError(w) }).
		ifSuccess(func() {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(commands)
		})
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
		return interactionResponseMessage("this interaction is too galaxy brain for me 😥"), true
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
	return interactionResponseMessage("boop")
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
	return interactionResponseMessage(content)
}

var InteractionResponseOof = interactionResponseMessage("oof please try again 😅")

func fuckoffInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	content, _ := foaas.FuckOff(fmt.Sprintf("<@%s>", interaction.Member.User.ID))
	if content == "" {
		return InteractionResponseOof
	}
	return interactionResponseMessage(content)
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

func when2meetInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
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
		fmt.Sprintf("<@%s> created a [when2meet](%s).", interaction.Member.User.ID, url))
}

func aboutmeInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	userId := interaction.Member.User.ID.String()
	for _, option := range interaction.Data.Options {
		switch option.Name {
		case "show":
			return showAboutme(ctx, userId)
		case "hide":
			return hideAboutme(ctx, userId)
		}
	}
	return InteractionResponseOof
}

func hideAboutme(ctx context.Context, userId string) discord.InteractionResponse {
	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		return InteractionResponseOof
	}
	if i, ok := leaders.GetIndex(userId); ok {
		leaders[i].Show = false
		if err := sheets.WriteLeaders(ctx, leaders); err != nil {
			logger.WithError(err).Error("sheets.WriteLeaders() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are hidden.")
	}
	return interactionResponseMessage("You dont have a blurb! :open_mouth:")
}

func interactionResponseMessage(text string) discord.InteractionResponse {
	return discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: text,
		},
	}
}

func showAboutme(ctx context.Context, userId string) discord.InteractionResponse {
	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		return InteractionResponseOof
	}

	if i, ok := leaders.GetIndex(userId); ok {
		leaders[i].Show = true
		if err := sheets.WriteLeaders(ctx, leaders); err != nil {
			logger.WithError(err).Error("sheets.WriteLeaders() failed")
			return InteractionResponseOof
		}
		return interactionResponseMessage(":person_gesturing_ok: You are visible.")
	}

	return interactionResponseMessage("You dont have a blurb! :open_mouth:")
}

func aboutmeCommandOptions(context.Context) ([]discord.ApplicationCommandOption, error) {
	return []discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "show",
			Description: "Show your information on the vvgo about us page.",
		},
		{
			Type:        discord.ApplicationCommandOptionTypeSubCommand,
			Name:        "hide",
			Description: "Hide your information on the vvgo about us page.",
		},
	}, nil
}
