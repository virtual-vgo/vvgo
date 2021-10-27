package slash_command

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discord "github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/clients/when2meet"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Handle))
	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{"type":1}`))
	require.NoError(t, err, "http.NewRequest() failed")
	req.Header.Set("X-Signature-Ed25519", "acbd")
	req.Header.Set("X-Signature-Timestamp", "1234")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "http.Do()")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "status code")
}

func TestHandleBeepInteraction(t *testing.T) {
	interaction := discord.Interaction{
		Type: discord.InteractionTypeApplicationCommand,
		Data: &discord.ApplicationCommandInteractionData{
			Name: "beep",
		},
	}
	response, ok := HandleInteraction(context.Background(), interaction)
	assert.True(t, ok)
	assertEqualInteractionResponse(t, discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{Content: "boop"},
	}, response)
}

func TestHandlePartsInteraction(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, redis.WriteSheet(ctx, models.SpreadsheetWebsiteData, models.SheetProjects, [][]interface{}{
		{"Name", "Title", "Parts Released"},
		{"10-hildas-healing", "Hilda's Healing", true},
	}))

	interaction := discord.Interaction{
		Type: discord.InteractionTypeApplicationCommand,
		Data: &discord.ApplicationCommandInteractionData{
			Name: "parts",
			Options: []discord.ApplicationCommandInteractionDataOption{
				{Name: "project", Value: "10-hildas-healing"},
			},
		},
	}

	response, ok := HandleInteraction(ctx, interaction)
	assert.True(t, ok)

	assertEqualInteractionResponse(t, discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Embeds: []discord.Embed{{
				Title:       "Hilda's Healing",
				Type:        "rich",
				Description: "· Parts are [here!](https://vvgo.org/parts?project=10-hildas-healing)\n· Submit files [here!]()\n· Submission Deadline: .",
				Url:         "https://vvgo.org/parts?project=10-hildas-healing",
				Color:       9181145,
				Footer:      &discord.EmbedFooter{Text: "Bottom text."},
			}},
		},
	}, response)
}

func TestHandleSubmissionInteraction(t *testing.T) {
	ctx := context.Background()
	redis.WriteSheet(ctx, models.SpreadsheetWebsiteData, models.SheetProjects, [][]interface{}{
		{"Name", "Title", "Parts Released", "Submission Link"},
		{"10-hildas-healing", "Hilda's Healing", true, "https://bit.ly/vvgo10submit"},
	})

	interaction := discord.Interaction{
		Type: discord.InteractionTypeApplicationCommand,
		Data: &discord.ApplicationCommandInteractionData{
			Name: "submit",
			Options: []discord.ApplicationCommandInteractionDataOption{
				{Name: "project", Value: "10-hildas-healing"},
			},
		},
	}

	response, ok := HandleInteraction(ctx, interaction)
	assert.True(t, ok)

	assertEqualInteractionResponse(t, discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Content: "[Submit here](https://bit.ly/vvgo10submit) for Hilda's Healing. Submission Deadline is ",
		},
	}, response)
}

func TestHandleWhen2MeetInteraction(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<body onload="window.location='/?10947260-c2u6i'">`))
	}))
	defer ts.Close()
	when2meet.Endpoint = ts.URL

	ctx := context.Background()
	interaction := discord.Interaction{
		Type:   discord.InteractionTypeApplicationCommand,
		Member: discord.GuildMember{User: discord.User{ID: "42069"}},
		Data: &discord.ApplicationCommandInteractionData{
			Name: "when2meet",
			Options: []discord.ApplicationCommandInteractionDataOption{
				{Name: "start_date", Value: "2030-02-01"},
				{Name: "end_date", Value: "2030-02-02"},
				{Name: "event_name", Value: "holy cheesus"},
			},
		},
	}

	response, ok := HandleInteraction(ctx, interaction)
	assert.True(t, ok)

	want := InteractionResponseMessage("<@42069> created a [when2meet](https://when2meet.com/?10947260-c2u6i).", true)
	assertEqualInteractionResponse(t, want, response)
}

func assertEqualInteractionResponse(t *testing.T, want, got discord.InteractionResponse) {
	assert.Equal(t, want.Type, got.Type, "interaction.Type")
	assertEqualInteractionApplicationCommandCallbackData(t, want.Data, got.Data)
}

func assertEqualInteractionApplicationCommandCallbackData(t *testing.T, want, got *discord.InteractionApplicationCommandCallbackData) {
	assert.Equal(t, want.Content, got.Content, "interaction.Data.HtmlSource")
	assert.Equal(t, want.TTS, got.TTS, "interaction.Data.TTS")
	assert.Equal(t, want.Embeds, got.Embeds, "interaction.Data.Embeds")
}
