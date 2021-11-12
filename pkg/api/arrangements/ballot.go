package arrangements

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"sort"
	"strings"
)

const Season = "season2"

type Ballot []string

func ServeBallot(r *http.Request) api.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)

	switch r.Method {
	case http.MethodGet:
		return handleGetBallot(ctx, identity)
	case http.MethodPost:
		return handlePostBallot(r, ctx, identity)
	default:
		return response.NewMethodNotAllowedError()
	}
}

func handleGetBallot(ctx context.Context, identity auth.Identity) api.Response {
	var ballotJSON string
	if err := redis.Do(ctx, redis.Cmd(&ballotJSON,
		"HGET", "arrangements:"+Season+":ballots", identity.DiscordID)); err != nil {
		logger.RedisFailure(ctx, err)
	}

	var ballot Ballot
	if ballotJSON != "" {
		if err := json.NewDecoder(strings.NewReader(ballotJSON)).Decode(&ballot); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			// Don't return here, instead we'll make a new ballot
		}
	}

	if len(ballot) == 0 {
		if err := redis.Do(ctx, redis.Cmd(&ballot,
			"LRANGE", "arrangements:"+Season+":submissions", "0", "-1")); err != nil {
			logger.RedisFailure(ctx, err)
			return response.NewInternalServerError()
		}
		sort.Strings(ballot)
	}

	return api.Response{Status: api.StatusOk, Ballot: ballot}
}

type PostBallotRequest []string

func handlePostBallot(r *http.Request, ctx context.Context, identity auth.Identity) api.Response {
	var ballot PostBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&ballot); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return response.NewJsonDecodeError(err)
	}
	if validateBallot(ctx, ballot) == false {
		return response.NewBadRequestError("invalid ballot")
	}

	ballotJSON, _ := json.Marshal(ballot)
	if err := redis.Do(ctx, redis.Cmd(nil,
		"HSET", "arrangements:"+Season+":ballots", identity.DiscordID, string(ballotJSON))); err != nil {
		logger.RedisFailure(ctx, err)
		return response.NewInternalServerError()
	}
	return api.NewOkResponse()
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
