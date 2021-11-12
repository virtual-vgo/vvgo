package slash_command

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/website_data"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
)

func SubmitCommandOptions(ctx context.Context) ([]discord.ApplicationCommandOption, error) {
	identity := auth.Anonymous()
	projects, err := website_data.ListProjects(ctx, identity)
	if err != nil {
		return nil, errors.ListProjectsFailure(err)
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
	identity := auth.Anonymous()
	projects, err := website_data.ListProjects(ctx, identity)
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
	} else if project, ok := projects.Get(projectName); ok {
		content = fmt.Sprintf(`[Submit here](%s) for %s. Submission Deadline is `, project.SubmissionLink, project.Title)
	}

	if content == "" {
		return InteractionResponseOof
	}
	return InteractionResponseMessage(content, true)
}
