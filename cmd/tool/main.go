package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"log"
	"net/http"
)

const ApplicationId = "700963768787795998"
const GuildId = "690626216637497425"
const RegistrationEndpoint = "https://discord.com/api/v8/applications/" + ApplicationId + "/guilds/" + GuildId + "/commands"

// https://discord.com/developers/docs/interactions/slash-commands#create-guild-application-command
type CreateApplicationCommand struct {
	Name        string                             `json:"name"`
	Description string                             `json:"description"`
	Options     []discord.ApplicationCommandOption `json:"options,omitempty"`
}

var beepCommand = CreateApplicationCommand{
	Name:        "beep",
	Description: "Send a beep",
}

func partsCommand() CreateApplicationCommand {
	identity := login.Anonymous()
	projects, err := sheets.ListProjects(context.Background(), &identity)
	if err != nil {
		log.Fatal("sheets.ListProjects() failed:", err)
	}
	projects = projects.Query(map[string]interface{}{
		"Hidden": false, "Video Released": false,
		"Parts Archived": false, "Parts Released": true})

	var choices []discord.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return CreateApplicationCommand{
		Name:        "parts",
		Description: "Link to parts",
		Options: []discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionTypeString,
				Name:        "project",
				Description: "Name of the project",
				Required:    true,
				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: "Hilda's Healing", Value: "10-hildas-healing"},
				},
			},
		},
	}
}

func main() {
	redis.InitializeFromEnv()
	client := discord.NewClient(context.Background())
	var authToken = client.Config.BotAuthToken

	registerCommand(authToken, beepCommand)
	registerCommand(authToken, partsCommand())
	listSlashCommands(authToken)
}

func registerCommand(AuthToken string, command CreateApplicationCommand) {
	var commandBytes bytes.Buffer
	if err := json.NewEncoder(&commandBytes).Encode(command); err != nil {
		log.Fatal("json.Encode() failed: ", err)
	}
	req, err := http.NewRequest(http.MethodPost, RegistrationEndpoint, &commandBytes)
	if err != nil {
		log.Fatal("http.NewRequest() failed: ", err)
	}
	req.Header.Set("Authorization", "Bot "+AuthToken)
	req.Header.Set("Content-Type", "application/json")
	doRequest(req)
}

func listSlashCommands(authToken string) {
	req, err := http.NewRequest(http.MethodGet, RegistrationEndpoint, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bot "+authToken)
	resp := doRequest(req)

	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	fmt.Println(body.String())
}

func doRequest(req *http.Request) *http.Response {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}
