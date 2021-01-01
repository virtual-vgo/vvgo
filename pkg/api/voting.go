package api

import (
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
)

type VotingView struct{}

func (x VotingView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ParseAndExecute(ctx, w, r, nil, "voting.gohtml")
}

type VotingCollector struct{
	GuildID discord.GuildID
}

func (x VotingCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := IdentityFromContext(ctx)
	var votes []string
	jsonDecode(r.Body, &votes)
	guildMember, _ := discord.NewClient(ctx).QueryGuildMember(ctx, x.GuildID, identity.DiscordID)
	redis.Do(ctx, redis.Cmd(nil, "SET", "votes:tester:"+identity.DiscordID.String()))
	logger.Printf("%s submitted votes: %s", guildMember.Nick, votes)
}
