package slash_command

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models"
)

func PartsCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := models.Anonymous()
	projects, err := models.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord.ApplicationCommandOption{ProjectCommandOption(projects.Current())}, nil
}

func PartsInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	identity := models.Anonymous()
	projects, err := models.ListProjects(ctx, &identity)
	if err != nil {
		logger.WithError(err).Error("sheets.ListProjects() failed")
		return InteractionResponseOof
	}

	project, ok := projects.Get(projectName)
	if !ok {
		return InteractionResponseOof
	}

	description := fmt.Sprintf(`· Parts are [here!](https://vvgo.org%s)
· Submit files [here!](%s)
· Submission Deadline: %s.`,
		project.PartsPage(), project.SubmissionLink, project.SubmissionDeadline)

	embed := discord.Embed{
		Title:       project.Title,
		Type:        discord.EmbedTypeRich,
		Description: description,
		Url:         "https://vvgo.org" + project.PartsPage(),
		Color:       0x8C17D9,
		Footer:      &discord.EmbedFooter{Text: "Bottom text."},
	}
	return discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageWithSource,
		Data: &discord.InteractionApplicationCommandCallbackData{
			Embeds: []discord.Embed{embed},
		},
	}
}
