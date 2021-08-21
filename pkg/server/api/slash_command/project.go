package slash_command

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models"
)

func ProjectCommandOption(projects models.Projects) discord.ApplicationCommandOption {
	var choices []discord.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return discord.ApplicationCommandOption{
		Type:        discord.ApplicationCommandOptionTypeString,
		Name:        "project",
		Description: "Name of the project",
		Required:    true,
		Choices:     choices,
	}
}
