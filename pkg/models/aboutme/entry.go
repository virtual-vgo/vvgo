package aboutme

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"strings"
)

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
			return nil, errors.RedisFailure(err)
		}
		dest := make(map[string]Entry)
		for _, entryJson := range buf {
			var entry Entry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, errors.JsonDecodeFailure(err)
			}
			dest[entry.DiscordID] = entry
		}
		return dest, nil
	} else {
		var buf []string
		cmd := "HMGET"
		args := append([]string{"about_me:entries"}, keys...)
		if err := redis.Do(ctx, redis.Cmd(&buf, cmd, args...)); err != nil {
			return nil, errors.RedisFailure(err)
		}
		dest := make(map[string]Entry)
		for _, entryJson := range buf {
			var entry Entry
			if err := json.NewDecoder(strings.NewReader(entryJson)).Decode(&entry); err != nil {
				return nil, errors.JsonDecodeFailure(err)
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
			return errors.JsonEncodeFailure(err)
		}
		args = append(args, id, buf.String())
	}

	if err := redis.Do(ctx, redis.Cmd(nil, "HMSET", args...)); err != nil {
		return errors.RedisFailure(err)
	}
	return nil
}

func DeleteEntries(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	args := append([]string{"about_me:entries"}, keys...)
	if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
		return errors.RedisFailure(err)
	}
	return nil
}
