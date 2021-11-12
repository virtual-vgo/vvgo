package channels

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"time"
)

var HandleList = cache.Handle(4*3600*time.Second, func(r *http.Request) api.Response {
	ctx := r.Context()
	channels, err := discord.GetGuildChannels(ctx)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(e); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return response.NewErrorResponse(response.Error{
				Code:    e.Code,
				Message: e.Error(),
				Data:    buf.Bytes(),
			})
		}
		return response.NewInternalServerError()
	}

	return api.Response{Status: api.StatusOk, Channels: channels}
})
