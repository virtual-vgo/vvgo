package api

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
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
			response = handleBeepInteraction()
		case "parts":
			response = handlePartsInteraction(interaction)
		}
	default:
		badRequest(w, "unsupported interaction type")
		return
	}
	handleError(json.NewEncoder(w).Encode(response)).logError("json.Encode() failed")
}

func handleBeepInteraction() discord.InteractionResponse {
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: "boop",
		},
	}
}

func handlePartsInteraction(interaction discord.Interaction) discord.InteractionResponse {
	var project discord.ApplicationCommandInteractionDataOption
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			project = option
		}
	}
	return discord.InteractionResponse{
		Type: discord.InteractionResponseTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: fmt.Sprintf("[%s](https://vvgo.org/parts?project=%s", project.Name, project.Value),
		},
	}
}
