package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"sort"
	"strings"
)

const season = "season2"

var VotingView = ServeTemplate("voting.gohtml")

var ArrangementsBallotApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := IdentityFromContext(ctx)

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="ballot.json"`)

		var ballotJSON string
		handleError(redis.Do(ctx, redis.Cmd(&ballotJSON, "HGET",
			"arrangements:"+season+":ballots", identity.DiscordID.String()))).
			logError("redis.Do() failed")
		if ballotJSON != "" {
			handleError(json.NewEncoder(w).Encode(json.RawMessage(ballotJSON))).
				logError("json.Encode() failed")
			return
		}

		var ballot []string
		handleError(redis.Do(ctx, redis.Cmd(&ballot, "LRANGE",
			"arrangements:"+season+":submissions", "0", "-1"))).
			logError("redis.Do() failed")
		sort.Strings(ballot)
		handleError(json.NewEncoder(w).Encode(ballot)).
			logError("json.Encode() failed")

	case http.MethodPost:
		var ballot []string

		handleError(json.NewDecoder(r.Body).Decode(&ballot)).
			logError("json.Decode() failed")
		if validateBallot(ctx, ballot) == false {
			badRequest(w, "invalid ballot")
			return
		}

		ballotJSON, _ := json.Marshal(ballot)
		handleError(redis.Do(ctx, redis.Cmd(nil, "HSET",
			"arrangements:"+season+":ballots", identity.DiscordID.String(), string(ballotJSON)))).
			logError("redis.Do() failed")
	}
})

func validateBallot(ctx context.Context, ballot []string) bool {
	var allowedChoices []string
	handleError(redis.Do(ctx, redis.Cmd(&allowedChoices, "LRANGE",
		"arrangements:"+season+":submissions", "0", "-1"))).
		logError("redis.Do() failed")

	if len(ballot) != len(allowedChoices) {
		return false
	}

	index := make(map[string]struct{}, len(allowedChoices))
	for _, choice := range allowedChoices {
		index[choice] = struct{}{}
	}

	for _, choice := range ballot {
		if _, ok := index[choice]; !ok {
			return false
		}
	}
	return true
}

var VotingResultsView = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	page := struct {
		Results []voteRanking
		Ballots []namedBallot
	}{
		Results: results,
		Ballots: nameBallots(ctx, data),
	}
	ParseAndExecute(ctx, w, r, &page, "voting_results.gohtml")
})

type namedBallot struct {
	Nick  string
	Votes []string
}

func nameBallots(ctx context.Context, data map[string]string) []namedBallot {
	ballots := make([]namedBallot, 0, len(data))
	for userID, ballotJSON := range data {
		var nick string
		guildMember, err := discord.NewClient(ctx).QueryGuildMember(ctx, discord.Snowflake(userID))
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
		sort.Strings(losers)
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
