package discord

// https://discord.com/developers/docs/interactions/slash-commands#interaction

type Interaction struct {
	ID        string                             `json:"id"`
	Type      InteractionType                    `json:"type"`
	Data      *ApplicationCommandInteractionData `json:"data"`
	GuildID   string                             `json:"guild_id"`
	ChannelID string                             `json:"channel_id"`
	Member    GuildMember                        `json:"member"`
	Token     string                             `json:"token"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-interactiontype

type InteractionType int

const (
	InteractionTypePing               InteractionType = 1
	InteractionTypeApplicationCommand InteractionType = 2
)

// https://discord.com/developers/docs/interactions/slash-commands#interaction-applicationcommandinteractiondataoption

type ApplicationCommandInteractionData struct {
	ID      string                                    `json:"id"`
	Name    string                                    `json:"name"`
	Options []ApplicationCommandInteractionDataOption `json:"options"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-applicationcommandinteractiondataoption

type ApplicationCommandInteractionDataOption struct {
	Name    string                                    `json:"name"`
	Value   string                                    `json:"value"`
	Options []ApplicationCommandInteractionDataOption `json:"options"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response

type InteractionResponse struct {
	Type InteractionResponseType                    `json:"type"`
	Data *InteractionApplicationCommandCallbackData `json:"data,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response-interactionresponsetype

type InteractionResponseType int

const (
	InteractionResponseTypePong                     InteractionResponseType = 1
	InteractionResponseTypeAcknowledge              InteractionResponseType = 2
	InteractionResponseTypeChannelMessage           InteractionResponseType = 3
	InteractionResponseTypeChannelMessageWithSource InteractionResponseType = 4
	InteractionResponseTypeAcknowledgeWithSource    InteractionResponseType = 5
)

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response-interactionapplicationcommandcallbackdata

type InteractionApplicationCommandCallbackData struct {
	TTS     bool   `json:"tts"`
	Content string `json:"content"`
}
