package arrangements

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"sort"
)

const Season = "season2"

func Ballot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")

		var ballotJSON string
		if err := redis.Do(ctx, redis.Cmd(&ballotJSON,
			"HGET", "arrangements:"+Season+":ballots", identity.DiscordID)); err != nil {
			logger.RedisFailure(ctx, err)
		}

		if ballotJSON != "" {
			if err := json.NewEncoder(w).Encode(json.RawMessage(ballotJSON)); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return
		}

		var ballot []string
		if err := redis.Do(ctx, redis.Cmd(&ballot,
			"LRANGE", "arrangements:"+Season+":submissions", "0", "-1")); err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.WriteInternalServerError(ctx, w)
			return
		}
		sort.Strings(ballot)

		if err := json.NewEncoder(w).Encode(ballot); err != nil {
			logger.JsonEncodeFailure(ctx, err)
		}
		return

	case http.MethodPost:
		var ballot []string

		if err := json.NewDecoder(r.Body).Decode(&ballot); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			http_helpers.WriteErrorBadRequest(ctx, w, "invalid json")
			return
		}

		if validateBallot(ctx, ballot) == false {
			http_helpers.WriteErrorBadRequest(ctx, w, "invalid ballot")
			return
		}

		ballotJSON, _ := json.Marshal(ballot)
		if err := redis.Do(ctx, redis.Cmd(nil,
			"HSET", "arrangements:"+Season+":ballots", identity.DiscordID, string(ballotJSON))); err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.WriteInternalServerError(ctx, w)
		}
		return
	}
}

func validateBallot(ctx context.Context, ballot []string) bool {
	var allowedChoices []string

	if err := redis.Do(ctx, redis.Cmd(&allowedChoices,
		"LRANGE", "arrangements:"+Season+":submissions", "0", "-1")); err != nil {
		logger.RedisFailure(ctx, err)
	}
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
