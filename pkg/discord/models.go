package discord

// https://discordapp.com/developers/docs/topics/oauth2
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// https://discordapp.com/developers/docs/reference#snowflakes
type Snowflake string

func (x Snowflake) String() string { return string(x) }

// https://discordapp.com/developers/docs/resources/user#user-object
type User struct {
	ID Snowflake `json:"id"`
}

// https://discordapp.com/developers/docs/resources/guild#guild-member-object
type GuildMember struct {
	Nick  string   `json:"nick"`
	Roles []string `json:"roles"`
}

// https://discord.com/developers/docs/interactions/slash-commands#create-guild-application-command
type CreateApplicationCommandParams struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Options     []ApplicationCommandOption `json:"options,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#applicationcommand
type ApplicationCommand struct {
	ID            string                     `json:"id"`
	ApplicationID string                     `json:"application_id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Options       []ApplicationCommandOption `json:"options,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#applicationcommandoption
type ApplicationCommandOption struct {
	Type        ApplicationCommandOptionType     `json:"type"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Required    bool                             `json:"required"`
	Choices     []ApplicationCommandOptionChoice `json:"choices,omitempty"`
	Options     []ApplicationCommandOption       `json:"options,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#applicationcommandoptiontype
type ApplicationCommandOptionType int

const (
	ApplicationCommandOptionTypeSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionTypeSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionTypeString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionTypeInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionTypeBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionTypeUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionTypeChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionTypeRole            ApplicationCommandOptionType = 8
)

// https://discord.com/developers/docs/interactions/slash-commands#applicationcommandoptionchoice
type ApplicationCommandOptionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

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
	Options []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-applicationcommandinteractiondataoption
type ApplicationCommandInteractionDataOption struct {
	Name    string                                    `json:"name"`
	Value   string                                    `json:"value"`
	Options []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
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
