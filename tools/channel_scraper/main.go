package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"
)

const endpoint = "https://discord.com/api/v8"
const guildId = "690626216637497425"
const limit = "100"
const outputDir = "scrape_out"

var ignoreChannelIds = []string{"690626217594060892", "817084492635701298"}

func isAllowedChannel(channel Channel) bool {
	if channel.Type != ChannelTypeGuildText {
		return false
	}

	for _, ignored := range ignoreChannelIds {
		if channel.Id == ignored {
			return false
		}
	}

	return true
}

var discordToken = os.Getenv("DISCORD_BOT_AUTHENTICATION_TOKEN")

const ChannelTypeGuildText = 0

var globalLimiter = func() <-chan struct{} {
	discordLimiter := time.Tick(time.Second * 1 / 40)
	cloudflareLimiter := time.Tick(time.Second * 10 * 60 / 9_000)
	ch := make(chan struct{})
	go func() {
		for {
			<-discordLimiter
			<-cloudflareLimiter
			ch <- struct{}{}
		}
	}()
	return ch
}()

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type Message struct {
	Id        string    `json:"id"`
	ChannelId string    `json:"channel_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type RateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

type Checkpoint struct {
	lock     sync.Mutex
	status   map[string]string
	fileName string
}

func ReadCheckpoint(fileName string) *Checkpoint {
	checkpointFile, err := os.Open(fileName)
	defer checkpointFile.Close()

	var status map[string]string
	switch {
	case err == nil:
		if err := json.NewDecoder(checkpointFile).Decode(&status); err != nil {
			log.Fatal(err)
		}
	case errors.Is(err, os.ErrNotExist):
		status = make(map[string]string)
	default:
		log.Fatal(err)
	}
	return &Checkpoint{status: status, fileName: fileName}
}

func (x *Checkpoint) Get(channelId string) string {
	x.lock.Lock()
	defer x.lock.Unlock()
	return x.status[channelId]
}

func (x *Checkpoint) Set(channelId, messageId string) {
	x.lock.Lock()
	defer x.lock.Unlock()
	x.status[channelId] = messageId

	file, err := os.OpenFile(x.fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(x.status); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	urlString := fmt.Sprintf(`%s/guilds/%s/channels`, endpoint, guildId)
	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bot "+discordToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := httputil.DumpResponse(resp, true)
		_, _ = fmt.Fprintln(os.Stdout, string(respBytes))
		log.Fatal("non-200 status code:", resp.StatusCode)
	}

	var channels []Channel
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		log.Fatal(err)
	}

	var wantChannels []Channel
	for _, channel := range channels {
		if isAllowedChannel(channel) {
			wantChannels = append(wantChannels, channel)
		}
	}

	checkpoint := ReadCheckpoint(path.Join(outputDir, "checkpoint.json"))
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(len(wantChannels))
	for _, channel := range wantChannels {
		fileName := path.Join(outputDir, fmt.Sprintf("channel_messages_%s.json", channel.Id))
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		msgEncoder := json.NewEncoder(file)

		go func(channel Channel) {
			defer wg.Done()
			defer file.Close()
			saveChannelMessages(ctx, checkpoint, channel, msgEncoder)
		}(channel)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	select {
	case sig := <-sigCh:
		log.Printf("caught signal %s", sig)
		cancel()
	}
	wg.Wait()
	log.Println("shutdown complete")
	os.Exit(0)
}

func saveChannelMessages(ctx context.Context, checkpoint *Checkpoint, channel Channel, msgEncoder *json.Encoder) {
	msgDataCh := fetchChannelMessages(ctx, channel.Id, checkpoint.Get(channel.Id))
	log.Printf("starting download for channel %s (%s)", channel.Name, channel.Id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("download interrupted for channel %s (%s)", channel.Name, channel.Id)
			return

		case data, ok := <-msgDataCh:
			if !ok {
				log.Printf("completed download for channel %s (%s)", channel.Name, channel.Id)
				return
			}

			for _, msg := range data {
				if err := msgEncoder.Encode(msg); err != nil {
					log.Println(err)
				}
			}
			checkpoint.Set(channel.Id, data[0].Id)
		}
	}
}

func fetchChannelMessages(ctx context.Context, channelId, beforeId string) <-chan []Message {
	output := make(chan []Message, 64)
	ready := make(chan struct{})
	readyAfter := func(d time.Duration) {
		go func() {
			if d != 0 {
				log.Printf("sleeping for %fs", d.Seconds())
				time.Sleep(d)
			}
			<-globalLimiter
			ready <- struct{}{}
		}()
	}
	readyAfter(0)

	go func() {
		defer close(output)
		for {
			select {
			case <-ready:
			case <-ctx.Done():
				return
			}

			params := make(url.Values)
			params.Set("limit", limit)
			if beforeId != "" {
				params.Set("before", beforeId)
			}

			urlString := fmt.Sprintf(`%s/channels/%s/messages?%s`, endpoint, channelId, params.Encode())
			req, err := http.NewRequest(http.MethodGet, urlString, nil)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Add("Authorization", "Bot "+discordToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			switch resp.StatusCode {
			case http.StatusOK:
				var data []Message
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					log.Fatal(err)
				}
				if len(data) == 0 || data[0].Id == beforeId {
					return
				}
				beforeId = data[0].Id
				output <- data
				readyAfter(0)

			case http.StatusTooManyRequests:
				var data RateLimitResponse
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					log.Fatal(err)
				}
				if data.RetryAfter == 0 {
					log.Fatal("too many rate-limits")
				}
				readyAfter(time.Duration(float64(time.Second) * data.RetryAfter))

			default:
				respBytes, _ := httputil.DumpResponse(resp, true)
				_, _ = fmt.Fprintln(os.Stdout, string(respBytes))
				log.Fatal("non-200 status code:", resp.StatusCode)
			}
		}
	}()
	return output
}
