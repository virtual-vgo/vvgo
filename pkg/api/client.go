package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	ClientConfig
}

type ClientConfig struct {
	ServerAddress string
	BasicAuthUser string
	BasicAuthPass string
}

func NewClient(config ClientConfig) *Client {
	return &Client{config}
}

func (x *Client) Upload(uploads ...Upload) ([]UploadStatus, error) {
	if len(uploads) == 0 {
		return []UploadStatus{}, nil
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(&uploads)
	req, err := http.NewRequest(http.MethodPost, x.ServerAddress+"/upload", &buffer)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Virtual-VGO Client")
	req.SetBasicAuth(x.BasicAuthUser, x.BasicAuthPass)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("http request received non-200 status: `%d: %s`", resp.StatusCode, bytes.TrimSpace(body))
	}

	var results []UploadStatus
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("json.Decode() failed: %v", err)
	}
	return results, nil
}
