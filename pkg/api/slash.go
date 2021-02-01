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

func SlashCommand(w http.ResponseWriter, r *http.Request) {
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
		ifSuccess(func() { handleInteraction(w, interaction) })
}

func handleInteraction(w http.ResponseWriter, interaction discord.Interaction) {
	var response discord.InteractionResponse
	switch interaction.Type {
	case discord.InteractionTypePing:
		response = discord.InteractionResponse{Type: discord.InteractionResponseTypePong}
	case discord.InteractionTypeApplicationCommand:
		switch interaction.Data.Name {
		case "beep":
			response = HandleBeepInteraction()
		case "parts":
			response = HandlePartsInteraction(interaction)
		default:
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

func HandleBeepInteraction() discord.InteractionResponse {
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: "boop",
		},
	}
}

func HandlePartsInteraction(interaction discord.Interaction) discord.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	var content string
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(context.Background(), &identity)
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

func HandleSubmissionInteraction(interaction discord.Interaction) discord.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	var content string
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(context.Background(), &identity)
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
