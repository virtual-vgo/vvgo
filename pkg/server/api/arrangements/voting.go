package arrangements

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"sort"
	"strings"
)

const Season = "season2"

func Ballot(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	switch r.Method {
	case http.MethodGet:
		return handleGetBallot(ctx, identity)
	case http.MethodPost:
		return handlePostBallot(r, ctx, identity)
	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

func handleGetBallot(ctx context.Context, identity models.Identity) models.ApiResponse {
	var ballotJSON string
	if err := redis.Do(ctx, redis.Cmd(&ballotJSON,
		"HGET", "arrangements:"+Season+":ballots", identity.DiscordID)); err != nil {
		logger.RedisFailure(ctx, err)
	}

	var ballot models.ArrangementsBallot
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
			return http_helpers.NewInternalServerError()
		}
		sort.Strings(ballot)
	}

	return models.ApiResponse{Status: models.StatusOk, Ballot: ballot}
}

type PostBallotRequest []string

func handlePostBallot(r *http.Request, ctx context.Context, identity models.Identity) models.ApiResponse {
	var ballot PostBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&ballot); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return http_helpers.NewJsonDecodeError(err)
	}
	if validateBallot(ctx, ballot) == false {
		return http_helpers.NewBadRequestError("invalid ballot")
	}

	ballotJSON, _ := json.Marshal(ballot)
	if err := redis.Do(ctx, redis.Cmd(nil,
		"HSET", "arrangements:"+Season+":ballots", identity.DiscordID, string(ballotJSON))); err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}
	return http_helpers.NewOkResponse()
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
