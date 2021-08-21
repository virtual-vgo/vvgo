package views

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
	"sort"
	"strings"
)

func VotingResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := make(map[string]string)
	if err := redis.Do(ctx, redis.Cmd(&data,
		"HGETALL", "arrangements:"+arrangements.Season+":ballots")); err != nil {
		logger.RedisFailure(ctx, err)
		helpers.InternalServerError(w)
		return
	}

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
}

type namedBallot struct {
	Nick  string
	Votes []string
}

func nameBallots(ctx context.Context, data map[string]string) []namedBallot {
	ballots := make([]namedBallot, 0, len(data))
	for userID, ballotJSON := range data {
		var nick string
		guildMember, err := discord.QueryGuildMember(ctx, discord.Snowflake(userID))
		if err != nil {
			logger.MethodFailure(ctx, "discord.QueryGuildMember", err)
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
