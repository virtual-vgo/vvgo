package slash_command

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models"
)

func SubmitCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := models.Anonymous()
	projects, err := models.ListProjects(ctx, &identity)
	if err != nil {
		return nil, fmt.Errorf("sheets.ListProjects() failed: %w", err)
	}
	return []discord.ApplicationCommandOption{ProjectCommandOption(projects.Current())}, nil
}

func SubmitInteractionHandler(ctx context.Context, interaction discord.Interaction) discord.InteractionResponse {
	var projectName string
	for _, option := range interaction.Data.Options {
		if option.Name == "project" {
			projectName = option.Value
		}
	}

	var content string
	identity := models.Anonymous()
	projects, err := models.ListProjects(ctx, &identity)
	if err != nil {
		logger.WithError(err).Error("sheets.ListProjects() failed")
	} else if project, ok := projects.Get(projectName); ok {
		content = fmt.Sprintf(`[Submit here](%s) for %s. Submission Deadline is `, project.SubmissionLink, project.Title)
	}

	if content == "" {
		return InteractionResponseOof
	}
	return InteractionResponseMessage(content, true)
}
