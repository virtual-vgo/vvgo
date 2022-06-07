package which_time

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"sort"
	"strings"
	"time"
)

const RedisKey = "cron:timezone_trumpet"

var locations []*time.Location

type CacheData struct {
	MessageId string
}

func init() {
	timezones := []string{
		"America/Los_Angeles",
		"America/New_York",
		"Europe/London",
	}

	for _, zone := range timezones {
		location, err := time.LoadLocation(zone)
		if err != nil {
			panic(err)
		}
		locations = append(locations, location)
	}
}

func WhichTime(ctx context.Context, channelId string) {

	var dataJson string
	if err := redis.Do(ctx, redis.Cmd(&dataJson, redis.HGET, RedisKey, channelId)); err != nil {
		logger.RedisFailure(ctx, err)
		return
	}

	var data CacheData
	if err := json.Unmarshal([]byte(dataJson), &data); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		// Do not return.
	}

	if data.MessageId != "" {
		editMessage(ctx, channelId, data.MessageId)
	} else {
		createMessage(ctx, channelId)
	}
}

func createMessage(ctx context.Context, channelId string) {
	message, err := discord.CreateMessage(ctx,
		discord.Snowflake(channelId),
		discord.CreateMessageParams{Embed: makeEmbed()},
	)

	if err != nil {
		logger.HttpDoFailure(ctx, err)
		return
	}

	data := CacheData{MessageId: message.Id}
	var dataJSON bytes.Buffer
	if err := json.NewEncoder(&dataJSON).Encode(&data); err != nil {
		logger.JsonEncodeFailure(ctx, err)
		return
	}

	if err := redis.Do(ctx, redis.Cmd(nil, redis.HSET, RedisKey, channelId, dataJSON.String())); err != nil {
		logger.RedisFailure(ctx, err)
		return
	}
}

func editMessage(ctx context.Context, channelId string, messageId string) {
	_, err := discord.EditMessage(ctx,
		discord.Snowflake(channelId),
		discord.Snowflake(messageId),
		discord.EditMessageParams{Embed: makeEmbed()},
	)

	if err != nil {
		logger.HttpDoFailure(ctx, err)
		if err := redis.Do(ctx, redis.Cmd(nil, redis.HDEL, RedisKey, channelId)); err != nil {
			logger.RedisFailure(ctx, err)
		}
	}
}

func makeEmbed() *discord.Embed {
	now := time.Now()
	times := make([]string, len(locations))
	for _, location := range locations {
		locationString := location.String()
		times = append(times, fmt.Sprintf("**%s:**\n> %s\n\n",
			strings.Replace(locationString[strings.LastIndex(locationString, "/")+1:], "_", " ", -1),
			now.In(location).Format(time.UnixDate),
		))
	}
	sort.Strings(times)
	return &discord.Embed{
		Title:       "🤔🤔 ¿¿ WHICH TIME IT IS ?? 🤔🤔",
		Description: "\n\n" + strings.Join(times, ""),
	}
}
