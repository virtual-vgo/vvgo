package slash_command

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
)

func BeepInteractionHandler(context.Context, discord.Interaction) discord.InteractionResponse {
	return InteractionResponseMessage("boop", false)
}
