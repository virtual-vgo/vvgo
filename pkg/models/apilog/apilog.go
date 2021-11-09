package apilog

import (
	"github.com/virtual-vgo/vvgo/pkg/version"
	"time"
)

const RedisKey = "logs:api"

type Entry struct {
	StartTime        time.Time       `json:"request_start"`
	RequestHost      string          `json:"request_host,omitempty"`
	RequestMethod    string          `json:"request_method,omitempty"`
	RequestBytes     int64           `json:"request_bytes,omitempty"`
	RequestUrl       string          `json:"request_url,omitempty"`
	RequestUserAgent string          `json:"request_user_agent,omitempty"`
	ResponseCode     int             `json:"response_code,omitempty"`
	ResponseBytes    int64           `json:"response_size,omitempty"`
	DurationSeconds  float64         `json:"duration_seconds,omitempty"`
	Version          version.Version `json:"version"`
}

func (entry Entry) Fields() map[string]interface{} {
	return map[string]interface{}{
		"start_time":       entry.StartTime,
		"request_host":     entry.RequestHost,
		"request_method":   entry.RequestMethod,
		"request_bytes":    entry.RequestBytes,
		"request_url":      entry.RequestUrl,
		"response_code":    entry.ResponseCode,
		"response_bytes":   entry.ResponseBytes,
		"duration_seconds": entry.DurationSeconds,
	}
}

type RedisQuery struct {
	StartTime       time.Time   `json:"start_time"`
	Cmd             interface{} `json:"cmd,omitempty"`
	ArgLen          interface{} `json:"arg_len,omitempty"`
	ArgBytes        interface{} `json:"arg_bytes,omitempty"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
}

func (entry RedisQuery) Fields() map[string]interface{} {
	return map[string]interface{}{
		"start_time":       entry.StartTime,
		"cmd":              entry.Cmd,
		"arg_len":          entry.ArgLen,
		"arg_bytes":        entry.ArgBytes,
		"duration_seconds": entry.DurationSeconds,
	}
}
