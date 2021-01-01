package api

import (
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
)

type VotingView struct{}

func (x VotingView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ParseAndExecute(ctx, w, r, nil, "voting.gohtml")
}

type VotingCollector struct{}

func (x VotingCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := IdentityFromContext(ctx)

	var config DiscordLoginConfig
	if err := parse_config.ReadFromRedisHash(ctx, "discord_login", &config); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
		internalServerError(w)
	}

	var votes []string
	jsonDecode(r.Body, &votes)
	guildMember, _ := discord.NewClient(ctx).QueryGuildMember(ctx, discord.GuildID(config.GuildID), identity.DiscordID)
	redis.Do(ctx, redis.Cmd(nil, "SET", "votes:tester:"+identity.DiscordID.String()))
	logger.Printf("%s submitted votes: %s", guildMember.Nick, votes)
}
