package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/http_util"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var ErrNon200Response = errors.New("non-200 response from discord")
var ErrInvalidOAuthCode = errors.New("invalid oauth code")

const ClientPublicKey = "a56a084a21829d02f272e4e3f4b67a846a831281849f6740f7bbf873840c4076"
const ApplicationID = "700963768787795998"
const OAuthClientID = ApplicationID
const VVGOGuildID = "690626216637497425" // The VVGO discord server
const VVGOVerifiedMemberRoleID = "690636730281230396"
const VVGOProductionTeamRoleID = "746434659252174971"
const VVGOExecutiveDirectorRoleID = "690626333062987866"
const VVGOProductionDirectorRoleID = "805504313072943155"

var IgnoreChannelIds = []string{
	"690626217594060892",
	"817084492635701298",
	"817084789139308544",
}

func LoginURL(state string) string {
	query := make(url.Values)
	query.Set("client_id", OAuthClientID)
	query.Set("redirect_uri", config.Env.VVGO.ServerUrl+"/login/discord")
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("scope", "identify")
	return "https://discord.com/api/oauth2/authorize?" + query.Encode()
}

// GetOAuthToken Query oauth token from discord.
// We use the authorization code grant.
func GetOAuthToken(ctx context.Context, code string) (*OAuthToken, error) {
	req, err := newOAuthRequest(ctx, code)
	if err != nil {
		return nil, err
	}

	var oauthToken OAuthToken
	if _, err := doDiscordRequest(req, &oauthToken); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// build the oauth request
func newOAuthRequest(ctx context.Context, code string) (*http.Request, error) {
	if code == "" {
		return nil, ErrInvalidOAuthCode
	}

	// build the authorization request
	form := make(url.Values)
	form.Add("client_id", OAuthClientID)
	form.Add("client_secret", config.Env.Discord.OAuthClientSecret)
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", config.Env.VVGO.ServerUrl+"/login/discord")
	form.Add("scope", "identify")

	req, err := newRequest(ctx, http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// GetIdentity Query discord for the token's identity.
// This requires an oauth token with identity scope.
// https://discordapp.com/developers/docs/resources/user#get-current-user
func GetIdentity(ctx context.Context, oauthToken *OAuthToken) (*User, error) {
	req, err := newTokenRequest(ctx, oauthToken, "/users/@me")
	if err != nil {
		return nil, err
	}

	var discordUser User
	if _, err := doDiscordRequest(req, &discordUser); err != nil {
		return nil, err
	}
	return &discordUser, nil
}

// returns a request using an oauth token for authentication
func newTokenRequest(ctx context.Context, oauthToken *OAuthToken, path string) (*http.Request, error) {
	// build the request
	req, err := newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.AccessToken))
	return req, err
}

// GetGuildMember Query discord for the guild member object of the guild id and user id.
// Here we use the server's own auth token.
// https://discordapp.com/developers/docs/resources/guild#get-guild-member
func GetGuildMember(ctx context.Context, userID Snowflake) (*GuildMember, error) {
	req, err := newBotRequest(ctx, http.MethodGet, "/guilds/"+VVGOGuildID+"/members/"+userID.String(), nil)
	if err != nil {
		logger.NewRequestFailure(ctx, err)
		return nil, err
	}

	// unmarshal the response
	var guildMember GuildMember
	if _, err := doDiscordRequest(req, &guildMember); err != nil {
		return nil, err
	}
	return &guildMember, nil
}

func SearchGuildMembers(ctx context.Context, query string, limit int) ([]GuildMember, error) {
	params := make(url.Values)
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))

	req, err := newBotRequest(ctx, http.MethodGet, "/guilds/"+VVGOGuildID+"/members", nil)
	if err != nil {
		logger.NewRequestFailure(ctx, err)
		return nil, err
	}

	// unmarshal the response
	var guildMembers []GuildMember
	if _, err := doDiscordRequest(req, &guildMembers); err != nil {
		return nil, err
	}
	return guildMembers, nil
}

func ListGuildMembers(ctx context.Context, limit int, after int) ([]GuildMember, error) {
	params := make(url.Values)
	if limit >= 1 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if after >= 0 {
		params.Set("after", strconv.Itoa(after))
	}

	req, err := newBotRequest(ctx, http.MethodGet, "/guilds/"+VVGOGuildID+"/members?"+params.Encode(), nil)
	if err != nil {
		logger.NewRequestFailure(ctx, err)
		return nil, err
	}

	// unmarshal the response
	var guildMembers []GuildMember
	if _, err := doDiscordRequest(req, &guildMembers); err != nil {
		return nil, err
	}
	return guildMembers, nil
}

func GetGuildChannels(ctx context.Context) ([]Channel, error) {
	const path = "/guilds/" + VVGOGuildID + "/channels"
	var dest []Channel
	err := doDiscordBotRequestWithJsonParams(ctx, http.MethodGet, path, nil, &dest)
	if err != nil {
		return nil, err
	}

	shouldSkip := func(id string) bool {
		for _, ignored := range IgnoreChannelIds {
			if id == ignored {
				return false
			}
		}
		return true
	}

	allowed := make([]Channel, len(dest))
	for _, ch := range dest {
		if shouldSkip(ch.Id) {
			continue
		}
		allowed = append(allowed, ch)
	}
	return dest, err
}

func CreateGuildChannel(ctx context.Context, params CreateGuildChannelParams) error {
	const path = "/guilds/" + VVGOGuildID + "/channels"
	return doDiscordBotRequestWithJsonParams(ctx, http.MethodPost, path, params, nil)
}

func GetApplicationCommands(ctx context.Context) ([]ApplicationCommand, error) {
	req, err := newSlashCommandRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	var commands []ApplicationCommand
	_, err = doDiscordRequest(req, &commands)
	return commands, err
}

func CreateApplicationCommand(ctx context.Context, params CreateApplicationCommandParams) (*ApplicationCommand, error) {
	var paramsBytes bytes.Buffer
	if err := json.NewEncoder(&paramsBytes).Encode(params); err != nil {
		return nil, fmt.Errorf("json.Encode() failed: %w", err)
	}

	req, err := newSlashCommandRequest(ctx, http.MethodPost, &paramsBytes)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var command ApplicationCommand
	_, err = doDiscordRequest(req, &command)
	if err != nil {
		return nil, err
	}
	return &command, nil
}

func DeleteApplicationCommand(ctx context.Context, id Snowflake) error {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands/" + id.String()
	req, err := newBotRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	_, err = doDiscordRequest(req, nil)
	return err
}

func CreateMessage(ctx context.Context, channelId Snowflake, params CreateMessageParams) (*Message, error) {
	path := "/channels/" + channelId.String() + "/messages"

	var message Message
	err := doDiscordBotRequestWithJsonParams(ctx, http.MethodPost, path, &params, &message)
	return &message, err
}

func EditMessage(ctx context.Context, channelId Snowflake, messageId Snowflake, params EditMessageParams) (*Message, error) {
	path := "/channels/" + channelId.String() + "/messages/" + messageId.String()

	var message Message
	err := doDiscordBotRequestWithJsonParams(ctx, http.MethodPatch, path, &params, &message)
	return &message, err
}

func BulkDeleteMessages(ctx context.Context, channelId Snowflake, params BulkDeleteMessagesParams) error {
	path := "/channels/" + channelId.String() + "/messages/bulk-delete"
	return doDiscordBotRequestWithJsonParams(ctx, http.MethodPost, path, &params, nil)
}

func newSlashCommandRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	path := "/applications/" + ApplicationID + "/guilds/" + VVGOGuildID + "/commands"
	return newBotRequest(ctx, method, path, body)
}

func doDiscordBotRequestWithJsonParams(ctx context.Context, method, path string, params interface{}, dest interface{}) error {
	var paramsBytes bytes.Buffer
	if err := json.NewEncoder(&paramsBytes).Encode(params); err != nil {
		return fmt.Errorf("json.Encode() failed: %w", err)
	}

	req, err := newBotRequest(ctx, method, path, &paramsBytes)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = doDiscordRequest(req, dest)
	return err
}

// returns a request using a bot token for authentication
func newBotRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bot "+config.Env.Discord.BotAuthenticationToken)
	return req, err
}

func newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, config.Env.Discord.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext() failed: %w", err)
	}
	return req, nil
}

// performs the http request and logs results
func doDiscordRequest(req *http.Request, dest interface{}) (*http.Response, error) {
	resp, err := http_util.DoRequest(req)
	if err != nil {
		logger.HttpDoFailure(req.Context(), err)
		return nil, err
	}

	switch {
	case resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent:
		logger.Info("discord api request complete")
		if dest != nil {
			err = json.NewDecoder(resp.Body).Decode(dest)
		}
		return resp, nil

	default:
		header := make(map[string]string)
		for key := range resp.Header {
			header[key] = resp.Header.Get(key)
		}

		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)

		var bodyJSON json.RawMessage
		if err := json.Unmarshal(body.Bytes(), &bodyJSON); err != nil {
			logger.JsonDecodeFailure(req.Context(), err)
			bodyJSON, _ = json.Marshal(body.String())
		}

		return nil, &Error{
			Code:   resp.StatusCode,
			Body:   bodyJSON,
			Header: header,
		}
	}
}

type Error struct {
	Code   int               `json:"code"`
	Header map[string]string `json:"header"`
	Body   json.RawMessage   `json:"body,omitempty"`
}

func (Error) Error() string { return "non-200 response from discord" }
