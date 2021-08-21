package slash_command

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discord2 "github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/clients/sheets"
	"github.com/virtual-vgo/vvgo/pkg/clients/when2meet"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/server/api/aboutme"
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
	interaction := discord2.Interaction{
		Type: discord2.InteractionTypeApplicationCommand,
		Data: &discord2.ApplicationCommandInteractionData{
			Name: "beep",
		},
	}
	response, ok := HandleInteraction(context.Background(), interaction)
	assert.True(t, ok)
	assertEqualInteractionResponse(t, discord2.InteractionResponse{
		Type: discord2.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord2.InteractionApplicationCommandCallbackData{Content: "boop"},
	}, response)
}

func TestHandlePartsInteraction(t *testing.T) {
	ctx := context.Background()
	sheets.WriteValuesToRedis(ctx, parse_config.Config.Sheets.WebsiteDataSpreadsheetID, "Projects", [][]interface{}{
		{"Name", "Title", "Parts Released"},
		{"10-hildas-healing", "Hilda's Healing", true},
	})

	interaction := discord2.Interaction{
		Type: discord2.InteractionTypeApplicationCommand,
		Data: &discord2.ApplicationCommandInteractionData{
			Name: "parts",
			Options: []discord2.ApplicationCommandInteractionDataOption{
				{Name: "project", Value: "10-hildas-healing"},
			},
		},
	}

	response, ok := HandleInteraction(ctx, interaction)
	assert.True(t, ok)

	assertEqualInteractionResponse(t, discord2.InteractionResponse{
		Type: discord2.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord2.InteractionApplicationCommandCallbackData{
			Embeds: []discord2.Embed{{
				Title:       "Hilda's Healing",
				Type:        "rich",
				Description: "· Parts are [here!](https://vvgo.org/parts?project=10-hildas-healing)\n· Submit files [here!]()\n· Submission Deadline: .",
				Url:         "https://vvgo.org/parts?project=10-hildas-healing",
				Color:       9181145,
				Footer:      &discord2.EmbedFooter{Text: "Bottom text."},
			}},
		},
	}, response)
}

func TestHandleSubmissionInteraction(t *testing.T) {
	ctx := context.Background()
	sheets.WriteValuesToRedis(ctx, parse_config.Config.Sheets.WebsiteDataSpreadsheetID, "Projects", [][]interface{}{
		{"Name", "Title", "Parts Released", "Submission Link"},
		{"10-hildas-healing", "Hilda's Healing", true, "https://bit.ly/vvgo10submit"},
	})

	interaction := discord2.Interaction{
		Type: discord2.InteractionTypeApplicationCommand,
		Data: &discord2.ApplicationCommandInteractionData{
			Name: "submit",
			Options: []discord2.ApplicationCommandInteractionDataOption{
				{Name: "project", Value: "10-hildas-healing"},
			},
		},
	}

	response, ok := HandleInteraction(ctx, interaction)
	assert.True(t, ok)

	assertEqualInteractionResponse(t, discord2.InteractionResponse{
		Type: discord2.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord2.InteractionApplicationCommandCallbackData{
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
	interaction := discord2.Interaction{
		Type:   discord2.InteractionTypeApplicationCommand,
		Member: discord2.GuildMember{User: discord2.User{ID: "42069"}},
		Data: &discord2.ApplicationCommandInteractionData{
			Name: "when2meet",
			Options: []discord2.ApplicationCommandInteractionDataOption{
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

func TestAboutmeHandler(t *testing.T) {
	ctx := context.Background()

	resetAboutMeEntries := func(t *testing.T) {
		require.NoError(t, redis.Do(ctx, redis.Cmd(nil, "DEL", "about_me:entries")))
	}

	aboutMeInteraction := func(cmd string, options []discord2.ApplicationCommandInteractionDataOption) discord2.Interaction {
		return discord2.Interaction{
			Type:   discord2.InteractionTypeApplicationCommand,
			Member: discord2.GuildMember{User: discord2.User{ID: "42069"}, Roles: []string{discord2.VVGOProductionTeamRoleID}},
			Data: &discord2.ApplicationCommandInteractionData{
				Name: "aboutme",
				Options: []discord2.ApplicationCommandInteractionDataOption{
					{Name: cmd, Options: options},
				},
			},
		}
	}

	testNotOnProductionTeam := func(t *testing.T, cmd string) {
		t.Run("not on production team", func(t *testing.T) {
			resetAboutMeEntries(t)

			interaction := aboutMeInteraction(cmd, nil)
			interaction.Member.Roles = nil
			response, ok := HandleInteraction(ctx, interaction)
			assert.True(t, ok)

			want := InteractionResponseMessage("Sorry, this tool is only for production teams. :bow:", true)
			assertEqualInteractionResponse(t, want, response)

			got, err := aboutme.ReadEntries(ctx, nil)
			assert.NoError(t, err)
			assert.Empty(t, got)
		})
	}

	t.Run("hide", func(t *testing.T) {
		testNotOnProductionTeam(t, "hide")

		t.Run("ok", func(t *testing.T) {
			resetAboutMeEntries(t)
			require.NoError(t, aboutme.WriteEntries(ctx,
				map[string]aboutme.Entry{"42069": {DiscordID: "42069", Show: true}}))

			response, ok := HandleInteraction(ctx, aboutMeInteraction("hide", nil))
			assert.True(t, ok)

			want := InteractionResponseMessage(":person_gesturing_ok: You are hidden from https://vvgo.org/about.", true)
			assertEqualInteractionResponse(t, want, response)

			got, err := aboutme.ReadEntries(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]aboutme.Entry{"42069": {DiscordID: "42069", Show: false}}, got)
		})

		t.Run("no blurb", func(t *testing.T) {
			resetAboutMeEntries(t)

			response, ok := HandleInteraction(ctx, aboutMeInteraction("hide", nil))
			assert.True(t, ok)

			want := InteractionResponseMessage("You dont have a blurb! :open_mouth:", true)
			assertEqualInteractionResponse(t, want, response)
		})
	})

	t.Run("show", func(t *testing.T) {
		testNotOnProductionTeam(t, "show")

		t.Run("ok", func(t *testing.T) {
			resetAboutMeEntries(t)
			require.NoError(t, aboutme.WriteEntries(ctx,
				map[string]aboutme.Entry{"42069": {DiscordID: "42069", Show: false}}))

			response, ok := HandleInteraction(ctx, aboutMeInteraction("show", nil))
			assert.True(t, ok)

			want := InteractionResponseMessage(":person_gesturing_ok: You are visible on https://vvgo.org/about.", true)
			assertEqualInteractionResponse(t, want, response)

			got, err := aboutme.ReadEntries(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]aboutme.Entry{"42069": {DiscordID: "42069", Show: true}}, got)
		})

		t.Run("no blurb", func(t *testing.T) {
			resetAboutMeEntries(t)
			response, ok := HandleInteraction(ctx, aboutMeInteraction("show", nil))
			assert.True(t, ok)

			want := InteractionResponseMessage("You dont have a blurb! :open_mouth:", true)
			assertEqualInteractionResponse(t, want, response)
		})
	})

	t.Run("update", func(t *testing.T) {
		testNotOnProductionTeam(t, "update")

		t.Run("exists", func(t *testing.T) {
			resetAboutMeEntries(t)
			require.NoError(t, aboutme.WriteEntries(ctx, map[string]aboutme.Entry{"42069": {DiscordID: "42069"}}))
			response, ok := HandleInteraction(ctx, aboutMeInteraction("update", []discord2.ApplicationCommandInteractionDataOption{
				{Name: "name", Value: "chester cheeta"},
				{Name: "blurb", Value: "dangerously cheesy"},
			}))
			assert.True(t, ok)

			want := InteractionResponseMessage(":person_gesturing_ok: It is written.", true)
			assertEqualInteractionResponse(t, want, response)

			got, err := aboutme.ReadEntries(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]aboutme.Entry{
				"42069": {DiscordID: "42069", Name: "chester cheeta", Title: "Production Team", Blurb: "dangerously cheesy"},
			}, got)
		})

		t.Run("doesnt exist", func(t *testing.T) {
			resetAboutMeEntries(t)
			response, ok := HandleInteraction(ctx, aboutMeInteraction("update", []discord2.ApplicationCommandInteractionDataOption{
				{Name: "name", Value: "chester cheeta"},
				{Name: "blurb", Value: "dangerously cheesy"},
			}))
			assert.True(t, ok)

			want := InteractionResponseMessage(":person_gesturing_ok: It is written.", true)
			assertEqualInteractionResponse(t, want, response)

			got, err := aboutme.ReadEntries(ctx, nil)
			assert.NoError(t, err)
			assert.Equal(t, map[string]aboutme.Entry{
				"42069": {DiscordID: "42069", Name: "chester cheeta", Title: "Production Team", Blurb: "dangerously cheesy"},
			}, got)
		})
	})
}

func assertEqualInteractionResponse(t *testing.T, want, got discord2.InteractionResponse) {
	assert.Equal(t, want.Type, got.Type, "interaction.Type")
	assertEqualInteractionApplicationCommandCallbackData(t, want.Data, got.Data)
}

func assertEqualInteractionApplicationCommandCallbackData(t *testing.T, want, got *discord2.InteractionApplicationCommandCallbackData) {
	assert.Equal(t, want.Content, got.Content, "interaction.Data.Content")
	assert.Equal(t, want.TTS, got.TTS, "interaction.Data.TTS")
	assert.Equal(t, want.Embeds, got.Embeds, "interaction.Data.Embeds")
}
