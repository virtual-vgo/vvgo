package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"log"
	"net/http"
	"os"
)

const ApplicationId = "700963768787795998"
const GuildId = "690626216637497425"
const RegistrationEndpoint = "https://discord.com/api/v8/applications/" + ApplicationId + "/guilds/" + GuildId + "/commands"

var beepCommand = discord.CreateApplicationCommandParams{
	Name:        "beep",
	Description: "Send a beep.",
}

func partsCommand(projects sheets.Projects) discord.CreateApplicationCommandParams {
	var choices []discord.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return discord.CreateApplicationCommandParams{
		Name:        "parts",
		Description: "Parts link for a project.",
		Options: []discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionTypeString,
				Name:        "project",
				Description: "Name of the project",
				Required:    true,
				Choices:     choices,
			},
		},
	}
}

func submissionCommand(projects sheets.Projects) discord.CreateApplicationCommandParams {
	var choices []discord.ApplicationCommandOptionChoice
	for _, project := range projects {
		choices = append(choices, discord.ApplicationCommandOptionChoice{
			Name: project.Title, Value: project.Name,
		})
	}
	return discord.CreateApplicationCommandParams{
		Name:        "submit",
		Description: "Submission link for a project.",
		Options: []discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionTypeString,
				Name:        "project",
				Description: "Name of the project",
				Required:    true,
				Choices:     choices,
			},
		},
	}
}

func main() {
	redis.InitializeFromEnv()
	client := discord.NewClient(context.Background())
	var authToken = client.Config.BotAuthenticationToken

	identity := login.Anonymous()
	currentProjects, err := sheets.ListProjects(context.Background(), &identity)
	if err != nil {
		log.Fatal("sheets.ListProjects() failed:", err)
	}
	currentProjects = currentProjects.Query(map[string]interface{}{
		"Hidden": false, "Video Released": false,
		"Parts Archived": false, "Parts Released": true})

	registerCommand(authToken, beepCommand)
	registerCommand(authToken, partsCommand(currentProjects))
	registerCommand(authToken, submissionCommand(currentProjects))
	listSlashCommands(authToken)
}

func registerCommand(AuthToken string, command discord.CreateApplicationCommandParams) {
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

	var body []discord.ApplicationCommand
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Println("json.Decode() failed: ", err)
	}
	outEncoder := json.NewEncoder(os.Stdout)
	outEncoder.SetIndent("", "  ")
	outEncoder.Encode(body)
}

func doRequest(req *http.Request) *http.Response {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}
