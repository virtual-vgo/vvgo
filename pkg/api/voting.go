package api

import (
	"encoding/json"
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

	var ballot []string
	jsonDecode(r.Body, &ballot) // decode ballot to make sure its valid json
	// re-encode to json
	ballotJSON, _ := json.Marshal(ballot)
	guildMember, _ := discord.NewClient(ctx).QueryGuildMember(ctx, discord.GuildID(config.GuildID), identity.DiscordID)
	redis.Do(ctx, redis.Cmd(nil, "HSET", "ballot:hash_tester", identity.DiscordID.String(), string(ballotJSON)))
	logger.Printf("%s submitted ballot: %s", guildMember.Nick, ballot)
}

type VotingResultsView struct{}

func (VotingResultsView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]string)
	redis.Do(ctx, redis.Cmd(&data, "HGETALL", "votes:hash_tester"))

	ballots := make([][]string, 0, len(data))
	for _, ballotJSON := range data {
		var ballot []string
		json.Unmarshal([]byte(ballotJSON), &ballot)
	}
	results := determineWinners(ballots)
	json.NewEncoder(w).Encode(results)
}

func determineWinners(ballots [][]string) []string {
	if len(ballots) < 1 {
		return []string{}
	}

	results := make([]string, 0, len(ballots[0]))
	for len(ballots[0]) > 0 {
		winner := pickNextWinner(ballots)
		results = append(results, winner)
		ballots = removeChoiceFromBallots(ballots, winner)
	}
	return results
}

func pickNextWinner(ballots [][]string) string {
	// score first choices
	firstChoice := make(map[string]int)

	for _, ballot := range ballots {
		firstChoice[ballot[0]] += 1
	}

	// find the max and min votes
	maxVotes := 0
	minVotes := len(ballots)
	for _, votes := range firstChoice {
		if votes > maxVotes {
			maxVotes = votes
		}
		if votes < minVotes {
			minVotes = votes
		}
	}

	// is max votes the majority?
	if float64(maxVotes) > float64(len(ballots))*0.5 {
		// find choice w/ max votes and return it
		for choice, votes := range firstChoice {
			if votes == maxVotes {
				return choice
			}
		}
	}

	// remove losers from the ballot
	var losers []string
	for choice, votes := range firstChoice {
		if votes == minVotes {
			losers = append(losers, choice)
		}
	}
	return pickNextWinner(removeChoiceFromBallots(ballots, losers...))
}

func removeChoiceFromBallots(ballots [][]string, choices ...string) [][]string {
	newBallots := make([][]string, len(ballots))
	for i := range ballots {
		var newChoices []string
		for _, ballotChoice := range ballots[i] {
			// should we keep this choice?
			keep := true
			for _, badChoice := range choices {
				if ballotChoice == badChoice {
					keep = false
				}
			}
			if keep {
				newChoices = append(newChoices, ballotChoice)
			}
		}
		newBallots[i] = newChoices
	}
	return newBallots
}
