package auth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strconv"
)

type OAuthRedirect struct {
	DiscordURL string
	State      string
	Secret     string
}

func ServeOAuthRedirect(r *http.Request) api.Response {
	ctx := r.Context()

	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.MethodFailure(ctx, "rand.Read", err)
		return response.NewInternalServerError()
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	secret := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "oauth_state:"+state, "300", secret)); err != nil {
		logger.RedisFailure(ctx, err)
		return response.NewInternalServerError()
	}

	return api.Response{Status: api.StatusOk, OAuthRedirect: &OAuthRedirect{
		DiscordURL: discord.LoginURL(state),
		State:      state,
		Secret:     secret,
	}}
}

func validateState(ctx context.Context, state, secret string) bool {
	var wantSecret string
	err := redis.Do(ctx, redis.Cmd(&wantSecret, "GET", "oauth_state:"+state))
	switch {
	case err != nil:
		logger.RedisFailure(ctx, err)
		return false
	case wantSecret == "":
		return false
	case secret == "":
		return false
	default:
		return secret == wantSecret
	}
}
