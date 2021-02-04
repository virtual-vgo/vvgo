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
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var SlashCommands = []SlashCommand{
	{
		Name:        "beep",
		Description: "Send a beep.",
		Handler:     BeepInteractionHandler,
	},
	{
		Name:        "parts",
		Description: "Parts link for a project.",
		Options:     PartsCommandOptions,
		Handler:     PartsInteractionHandler,
	},
	{
		Name:        "submit",
		Description: "Submission link for a project.",
		Options:     SubmissionCommandOptions,
		Handler:     SubmissionInteractionHandler,
	},
}

func CreateSlashCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	for _, command := range SlashCommands {
		handleError(command.Create(ctx)).logError("SlashCommand.Create() failed")
	}
	ViewSlashCommands(w, r)
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
	handleError(json.NewDecoder(&body).Decode(&interaction)).
		logError("json.Decode() failed").
		ifError(func(err error) { badRequest(w, "invalid request body: "+err.Error()) }).
		ifSuccess(func() { handleInteraction(w, ctx, interaction) })
}

func handleInteraction(w http.ResponseWriter, ctx context.Context, interaction discord.Interaction) {
	var response discord.InteractionResponse
	switch interaction.Type {
	case discord.InteractionTypePing:
		response = discord.InteractionResponse{Type: discord.InteractionResponseTypePong}
	case discord.InteractionTypeApplicationCommand:
		var handler InteractionHandler
		for _, command := range SlashCommands {
			if interaction.Data.Name == command.Name {
				handler = command.Handler
			}
		}
		if handler != nil {
			response = handler(ctx, interaction)
		} else {
			response = discord.InteractionResponse{
				Type: discord.InteractionResponseTypeChannelMessageWithSource,
				Data: &discord.InteractionApplicationCommandCallbackData{
					Content: fmt.Sprintf("i don't how to %s yet 😥", interaction.Data.Name)}}
		}
	default:
		badRequest(w, "unsupported interaction type")
		return
	}
	handleError(json.NewEncoder(w).Encode(response)).logError("json.Encode() failed")
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

func BeepInteractionHandler(context.Context, discord.Interaction) discord.InteractionResponse {
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: "boop",
		},
	}
}

func PartsCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord.ApplicationCommandOption{projectCommandOption(projects.Current())}, nil
}

func PartsInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
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
		content = fmt.Sprintf("[Parts for %s](https://vvgo.org%s)", project.Title, project.PartsPage())
	}

	if content == "" {
		content = "oof please try again 😅"
	}
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessage,
		Data: &discord.InteractionApplicationCommandCallbackData{Content: content},
	}
}

func SubmissionCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
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

func SubmissionInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
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
		content = fmt.Sprintf("[Submit here for %s](%s)", project.Title, project.SubmissionLink)
	}

	if content == "" {
		content = "oof please try again 😅"
	}
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessage,
		Data: &discord.InteractionApplicationCommandCallbackData{Content: content},
	}
}
