package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"sort"
	"strings"
)

type VotingView struct{}

const season = "season2"

func (x VotingView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var choices []string
	redis.Do(ctx, redis.Cmd(&choices, "LRANGE", "arrangements:"+season+":submissions", "0", "-1"))
	sort.Strings(choices)
	ParseAndExecute(ctx, w, r, &choices, "voting.gohtml")
}

type arrangementSubmission struct {
	PieceTitle     string `json:"piece_title"`
	SourceMaterial string `json:"source_material"`
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
	redis.Do(ctx, redis.Cmd(nil, "HSET", "arrangements:"+season+":ballots", identity.DiscordID.String(), string(ballotJSON)))
	logger.Printf("%s submitted ballot: %s", guildMember.Nick, ballot)
}

type VotingResultsView struct{}

func (VotingResultsView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]string)
	redis.Do(ctx, redis.Cmd(&data, "HGETALL", "arrangements:"+season+":ballots"))

	ballots := make([][]string, 0, len(data))
	for _, ballotJSON := range data {
		var ballot []string
		_ = json.Unmarshal([]byte(ballotJSON), &ballot)
		ballots = append(ballots, ballot)
	}
	results := rankResultsForView(determineWinners(ballots))

	var config DiscordLoginConfig
	if err := parse_config.ReadFromRedisHash(ctx, "discord_login", &config); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
		internalServerError(w)
	}

	page := struct {
		Results []voteRanking
		Ballots []namedBallot
	}{
		Results: results,
		Ballots: nameBallots(ctx, discord.GuildID(config.GuildID), data),
	}
	ParseAndExecute(ctx, w, r, &page, "voting_results.gohtml")
}

type namedBallot struct {
	Nick  string
	Votes []string
}

func nameBallots(ctx context.Context, guildId discord.GuildID, data map[string]string) []namedBallot {
	ballots := make([]namedBallot, 0, len(data))
	for userID, ballotJSON := range data {
		var nick string
		guildMember, err := discord.NewClient(ctx).QueryGuildMember(ctx, guildId, discord.UserID(userID))
		if err != nil {
			logger.WithError(err).Error("discord.QueryGuildMember() failed")
		} else {
			nick = guildMember.Nick
		}

		var votes []string
		_ = json.Unmarshal([]byte(ballotJSON), &votes)
		ballots = append(ballots, namedBallot{Votes: votes, Nick: nick})
	}
	return ballots
}

func rankResults(results [][]string) map[int][]string {
	rank := make(map[int][]string)
	count := 1
	for _, row := range results {
		rank[count] = row
		count += len(row)
	}
	return rank
}

type voteRanking struct {
	Rank  int
	Names string
}

func rankResultsForView(results [][]string) []voteRanking {
	ranking := make([]voteRanking, len(results))
	count := 1
	for i := range results {
		ranking[i].Names = strings.Join(results[i], ", ")
		ranking[i].Rank = count
		count += len(results[i])
	}
	return ranking
}

func determineWinners(ballots [][]string) [][]string {
	if len(ballots) < 1 {
		return [][]string{}
	}

	results := make([][]string, 0, len(ballots[0]))
	for len(ballots[0]) > 0 {
		winner := pickNextWinner(ballots)
		results = append(results, winner)
		ballots = removeChoiceFromBallots(ballots, winner...)
	}
	return results
}

func pickNextWinner(ballots [][]string) []string {
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
				return []string{choice}
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

	// dead tie
	if minVotes == maxVotes {
		return losers
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
