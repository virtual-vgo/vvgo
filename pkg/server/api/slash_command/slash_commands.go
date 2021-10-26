package slash_command

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
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
		Options:     SubmitCommandOptions,
		Handler:     SubmitInteractionHandler,
	},
	{
		Name:        "fuckoff",
		Description: "A modern solution to the common problem of telling people to fuck off.",
		Handler:     FuckoffInteractionHandler,
	},
	{
		Name:        "when2meet",
		Description: "Make a when2meet link.",
		Options:     When2meetCommandOptions,
		Handler:     when2meetInteractionHandler,
	},
	{
		Name:        "aboutme",
		Description: "Manage your about me blurb on the vvgo website.",
		Options:     AboutmeCommandOptions,
		Handler:     AboutmeInteractionHandler,
	},
}

var InteractionResponseOof = InteractionResponseMessage("oof please try again 😅", true)
var InteractionResponseGalaxyBrain = InteractionResponseMessage("this interaction is too galaxy brain for me 😥", true)

func Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	for _, command := range SlashCommands {
		<-timer.C
		if err := command.Create(ctx); err != nil {
			logger.MethodFailure(ctx, "SlashCommand.Create", err)
			http_helpers.WriteInternalServerError(ctx, w)
			return
		} else {
			logger.Info(command.Name, "command created")
		}
	}
	http.Redirect(w, r, "/slash_commands", http.StatusFound)
}

func List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commands, err := discord.GetApplicationCommands(ctx)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetApplicationCommands", err)
		http_helpers.WriteInternalServerError(ctx, w)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commands); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body bytes.Buffer
	_, _ = body.ReadFrom(r.Body)

	publicKey, _ := hex.DecodeString(discord.ClientPublicKey)
	if len(publicKey) == 0 {
		logger.Error("invalid discord public key")
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	signature, _ := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
	if len(signature) == 0 {
		http_helpers.WriteErrorBadRequest(ctx, w, "invalid signature")
		return
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	if len(timestamp) == 0 {
		http_helpers.WriteErrorBadRequest(ctx, w, "invalid signature timestamp")
		return
	}

	if ed25519.Verify(publicKey, []byte(timestamp+body.String()), signature) == false {
		http_helpers.WriteUnauthorizedError(ctx, w)
		return
	}

	var interaction discord.Interaction
	if err := json.NewDecoder(&body).Decode(&interaction); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		http_helpers.WriteErrorBadRequest(ctx, w, "invalid request body: "+err.Error())
		return
	}

	response, ok := HandleInteraction(ctx, interaction)
	if !ok {
		http_helpers.WriteErrorBadRequest(ctx, w, "unsupported interaction type")
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.JsonEncodeFailure(ctx, err)
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
	_, err = discord.CreateApplicationCommand(ctx, params)
	return err
}

func InteractionResponseMessage(text string, ephemeral bool) discord.InteractionResponse {
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
