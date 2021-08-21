package foaas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"math/rand"
	"net/http"
	"strings"
)

const Endpoint = "https://foaas.com"
const OperationsEndpoint = "https://foaas.com/operations"

type Operation struct {
	Name   string
	Url    string
	Fields []struct {
		Name  string
		Field string
	}
}

type Response struct {
	Message  string `json:"message"`
	Subtitle string `json:"subtitle"`
}

func FuckOff(from string) (string, error) {
	resp, err := http_wrappers.Get(OperationsEndpoint)
	if err != nil {
		return "", fmt.Errorf("http.Do() failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	var operations []Operation
	if err := json.NewDecoder(resp.Body).Decode(&operations); err != nil {
		return "", fmt.Errorf("json.Decode() failed %w", err)
	}

	var usable []Operation
	for _, operation := range operations {
		if len(operation.Fields) != 1 {
			continue
		}
		if operation.Fields[0].Field != "from" {
			continue
		}
		usable = append(usable, operation)
	}

	if len(usable) == 0 {
		return "", errors.New("no usable operation")
	}

	operation := usable[rand.Int()%len(usable)]
	url := Endpoint + strings.Replace(operation.Url, ":from", from, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("http.NewRequest() failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err = http_wrappers.DoRequest(req)
	if err != nil {
		return "", fmt.Errorf("http.Do() failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	var respObject Response
	if err := json.NewDecoder(resp.Body).Decode(&respObject); err != nil {
		return "", fmt.Errorf("json.Decode() failed %w", err)
	}
	return respObject.Message + " " + respObject.Subtitle, nil
}
