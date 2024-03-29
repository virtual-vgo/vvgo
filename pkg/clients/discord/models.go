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
	ID       Snowflake `json:"id"`
	Username string    `json:"username"`
}

// https://discordapp.com/developers/docs/resources/guild#guild-member-object
type GuildMember struct {
	User  User     `json:"user"`
	Nick  string   `json:"nick"`
	Roles []string `json:"roles"`
}

// https://discord.com/developers/docs/resources/channel#edit-message-jsonform-params
type CreateMessageParams struct {
	Content string `json:"content,omitempty"`
	Embed   *Embed `json:"embed,omitempty"`
}

type EditMessageParams CreateMessageParams

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
	Type InteractionCallbackType                    `json:"type"`
	Data *InteractionApplicationCommandCallbackData `json:"data,omitempty"`
}

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response-interactioncallbacktype
type InteractionCallbackType int

const (
	InteractionCallbackTypePong                     InteractionCallbackType = 1
	InteractionCallbackTypeChannelMessageWithSource InteractionCallbackType = 4
	InteractionCallbackTypeAcknowledgeWithSource    InteractionCallbackType = 5
)

// https://discord.com/developers/docs/interactions/slash-commands#interaction-response-interactionapplicationcommandcallbackdata
type InteractionApplicationCommandCallbackData struct {
	TTS     bool    `json:"tts"`
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds,omitempty"`
	Flags   int     `json:"flags,omitempty"`
}

const InteractionApplicationCommandCallbackDataFlagEphemeral = 1 << 6

// https://discord.com/developers/docs/resources/webhook#execute-webhook-jsonform-params
type ExecuteWebhookParams struct {
	Content   string  `json:"content,omitempty"`
	Username  string  `json:"username,omitempty"`
	AvatarUrl string  `json:"avatar_url,omitempty"`
	TTS       bool    `json:"tts"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-structure
type Embed struct {
	Title       string          `json:"title,omitempty"`
	Type        EmbedType       `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	Url         string          `json:"url,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Video       *EmbedVideo     `json:"video,omitempty"`
	Provider    *EmbedProvider  `json:"provider,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []EmbedField    `json:"fields,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-types
type EmbedType string

const (
	EmbedTypeRich    EmbedType = "rich"
	EmbedTypeImage   EmbedType = "image"
	EmbedTypeVideo   EmbedType = "video"
	EmbedTypeGifv    EmbedType = "gifv"
	EmbedTypeArticle EmbedType = "article"
	EmbedTypeLink    EmbedType = "link"
)

// https://discord.com/developers/docs/resources/channel#embed-object-embed-thumbnail-structure
type EmbedThumbnail struct {
	Url      string `json:"url,omitempty"`
	ProxyUrl string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-video-structure
type EmbedVideo struct {
	Url      string `json:"url,omitempty"`
	ProxyUrl string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-image-structure
type EmbedImage struct {
	Url      string `json:"url,omitempty"`
	ProxyUrl string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-provider-structure
type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-author-structure
type EmbedAuthor struct {
	Name         string `json:"name,omitempty"`
	Url          string `json:"url,omitempty"`
	IconUrl      string `json:"icon_url,omitempty"`
	ProxyIconUrl string `json:"proxy_icon_url,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-footer-structure
type EmbedFooter struct {
	Text         string `json:"text,omitempty"`
	IconUrl      string `json:"icon_url,omitempty"`
	ProxyIconUrl string `json:"proxy_icon_url,omitempty"`
}

// https://discord.com/developers/docs/resources/channel#embed-object-embed-field-structure
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// https://discord.com/developers/docs/resources/channel#message-object
type Message struct {
	Id        string `json:"id"`
	ChannelId string `json:"channel_id"`
	Content   string `json:"content"`
}

type BulkDeleteMessagesParams struct {
	Messages []string `json:"messages"`
}

// https://discord.com/developers/docs/resources/channel#channel-object-channel-types
type ChannelType int

const (
	ChannelTypeGuildText     ChannelType = 0
	ChannelTypeDM            ChannelType = 1
	ChannelTypeGuildVoice    ChannelType = 2
	ChannelTypeGroupDM       ChannelType = 3
	ChannelTypeGuildCategory ChannelType = 4
)

type CreateGuildChannelParams struct {
	Name  string      `json:"name"`
	Type  ChannelType `json:"type,omitempty"`
	Topic string      `json:"topic,omitempty"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type RateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}
