package api

import (
	"encoding/json"
	"net/http"
)

// https://discord.com/developers/docs/interactions/slash-commands#interaction
const (
	InteractionTypePing               = 1
	InteractionTypeApplicationCommand = 2
)

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response-interactionresponsetype
const (
	InteractionResponseTypeChannelMessageWithSource = 4
)

var PartsCommand = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var request map[string]interface{}
	handleError(json.NewDecoder(r.Body).Decode(&request)).logError("json.Decode() failed").
		ifError(func(err error) { badRequest(w, "invalid request body: "+err.Error()) }).
		ifSuccess(func() {
			response := map[string]interface{}{
				"type": InteractionResponseTypeChannelMessageWithSource,
				"data": map[string]interface{}{"tts": false, "content": "https://vvgo.org/parts"},
			}
			handleError(json.NewEncoder(w).Encode(response)).logError("json.Encode() failed")
		})
})
