package channels

import (
	"bytes"
	"encoding/json"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"time"
)

var ServeChannels = cache.Handle(4*3600*time.Second, func(r *http.Request) http2.Response {
	ctx := r.Context()
	channels, err := discord.GetGuildChannels(ctx)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(e); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return errors.NewErrorResponse(errors.Error{
				Code:    e.Code,
				Message: e.Error(),
				Data:    buf.Bytes(),
			})
		}
		return errors.NewInternalServerError()
	}

	return http2.Response{Status: http2.StatusOk, Channels: channels}
})
