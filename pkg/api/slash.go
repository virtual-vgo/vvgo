package api

// https://discord.com/developers/docs/interactions/slash-commands

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"net/http"
)

func SlashCommand(w http.ResponseWriter, r *http.Request) {
	var request discord.Interaction
	handleError(json.NewDecoder(r.Body).Decode(&request)).logError("json.Decode() failed").
		ifError(func(err error) { badRequest(w, "invalid request body: "+err.Error()) }).
		ifSuccess(func() {
			var response discord.InteractionResponse
			switch request.Type {
			case discord.InteractionTypePing:
				response = discord.InteractionResponse{Type: discord.InteractionResponseTypePong}
			case discord.InteractionTypeApplicationCommand:
				switch request.Data.Name {
				case "beep":
					response = discord.InteractionResponse{
						Type: discord.InteractionResponseTypeChannelMessageWithSource,
						Data: discord.InteractionApplicationCommandCallbackData{
							TTS:     false,
							Content: "boop",
						},
					}
				}
			default:
				badRequest(w, "unsupported interaction type")
				return
			}
			handleError(json.NewEncoder(w).Encode(response)).logError("json.Encode() failed")
		})
}
