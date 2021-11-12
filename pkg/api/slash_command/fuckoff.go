package slash_command

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/foaas"
)

func FuckoffInteractionHandler(_ context.Context, interaction discord.Interaction) discord.InteractionResponse {
	content, _ := foaas.FuckOff(fmt.Sprintf("<@%s>", interaction.Member.User.ID))
	if content == "" {
		return InteractionResponseOof
	}
	return InteractionResponseMessage(content, true)
}
