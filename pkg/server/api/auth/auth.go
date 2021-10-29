package auth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strconv"
	"time"
)

const (
	SessionDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks
)

func OAuthRedirect(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.MethodFailure(ctx, "rand.Read", err)
		return http_helpers.NewInternalServerError()
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	secret := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "oauth_state:"+state, "300", secret)); err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	return models.ApiResponse{Status: models.StatusOk, OAuthRedirect: &models.OAuthRedirect{
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
