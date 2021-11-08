package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/models/apilog"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strings"
	"time"
)

func LogsHandler(r *http.Request) models.ApiResponse {
	params := r.URL.Query()
	start, _ := time.Parse(time.RFC3339, params.Get("start"))
	end, _ := time.Parse(time.RFC3339, params.Get("end"))

	if start.IsZero() {
		start = time.Now().Add(-52 * 7 * 24 * 3600 * time.Second)
	}

	if end.IsZero() {
		end = time.Now()
	}

	logsJSON, err := redis.ListLogs(r.Context(), start, end)
	if err != nil {
		logger.RedisFailure(r.Context(), err)
		return http_helpers.NewRedisError(err)
	}

	entries := make([]apilog.Entry, 0, len(logsJSON))
	for _, logJSON := range logsJSON {
		var entry apilog.Entry
		if err := json.NewDecoder(strings.NewReader(logJSON)).Decode(&entry); err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
		entries = append(entries, entry)
	}

	return models.ApiResponse{Status: models.StatusOk, ApiLogs: entries}
}
