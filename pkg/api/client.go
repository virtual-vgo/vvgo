package api

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
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
	gob.NewEncoder(&buffer).Encode(&uploads)
	req, err := x.newRequest(http.MethodPost, x.ServerAddress+"/upload", &buffer)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
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
	if err := gob.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("json.Decode() failed: %v", err)
	}
	return results, nil
}

func (x *Client) Authenticate() error {
	req, err := x.newRequest(http.MethodGet, x.ServerAddress+"/auth", strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("http.NewRequest() failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do() failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-200 status `%d: %s`", resp.StatusCode, bytes.TrimSpace(buf))
	}
	return nil
}

func (x *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "Virtual-VGO Client")
	req.SetBasicAuth(x.BasicAuthUser, x.BasicAuthPass)
	return req, nil
}
