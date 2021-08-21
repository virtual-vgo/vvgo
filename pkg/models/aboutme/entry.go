package aboutme

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/error_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"strings"
)

var logger = log.New()

type Entry struct {
	DiscordID string `json:"discord_id,omitempty"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Blurb     string `json:"blurb"`
	Show      bool   `json:"show"`
}

func ReadEntries(ctx context.Context, keys []string) (map[string]Entry, error) {
	if keys == nil {
		buf := make(map[string]string)
		cmd := "HGETALL"
		args := []string{"about_me:entries"}
		if err := redis.Do(ctx, redis.Cmd(&buf, cmd, args...)); err != nil {
			return nil, error_wrappers.RedisFailed(err)
		}
		dest := make(map[string]Entry)
		for _, entryJson := range buf {
			var entry Entry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, error_wrappers.JsonDecodeFailed(err)
			}
			dest[entry.DiscordID] = entry
		}
		return dest, nil
	} else {
		var buf []string
		cmd := "HMGET"
		args := append([]string{"about_me:entries"}, keys...)
		if err := redis.Do(ctx, redis.Cmd(&buf, cmd, args...)); err != nil {
			return nil, error_wrappers.RedisFailed(err)
		}
		dest := make(map[string]Entry)
		for _, entryJson := range buf {
			var entry Entry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, error_wrappers.JsonDecodeFailed(err)
			}
			dest[entry.DiscordID] = entry
		}
		return dest, nil
	}
}

func WriteEntries(ctx context.Context, src map[string]Entry) error {
	if len(src) == 0 {
		logger.Warnf("WriteEntries: there's nothing to write!")
		return nil
	}

	args := []string{"about_me:entries"}
	for id, entry := range src {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(entry); err != nil {
			return error_wrappers.JsonEncodeFailed(err)
		}
		args = append(args, id, buf.String())
	}

	if err := redis.Do(ctx, redis.Cmd(nil, "HMSET", args...)); err != nil {
		return error_wrappers.RedisFailed(err)
	}
	return nil
}

func DeleteEntries(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	args := append([]string{"about_me:entries"}, keys...)
	if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
		return error_wrappers.RedisFailed(err)
	}
	return nil
}
