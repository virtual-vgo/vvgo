package api

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const ClientUserAgent = "Virtual-VGO Client"

type Client struct {
	ClientConfig
}

type ClientConfig struct {
	ServerAddress string
	Token         string
}

func NewClient(config ClientConfig) *Client {
	return &Client{config}
}

// Backup sends a request to the server to make a backup
func (x *Client) Backup() error {
	form := make(url.Values)
	form.Set("cmd", "backup")
	req, err := x.newRequest(http.MethodPost, x.ServerAddress+"/backups?cmd=backup", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("received non-200 status: %d", resp.StatusCode)
	}
}

// GetProject queries the api server for the project data.
func (x *Client) GetProject(name string) (*projects.Project, error) {
	values := url.Values{"name": []string{name}}
	req, err := x.newRequest(http.MethodGet, x.ServerAddress+"/projects?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var project projects.Project
		if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
			return nil, err
		}
		return &project, nil
	case http.StatusNotFound:
		return nil, projects.ErrNotFound
	default:
		return nil, fmt.Errorf("received non-200 status: %d", resp.StatusCode)
	}
}

func (x *Client) Upload(uploads ...Upload) []UploadStatus {
	if len(uploads) == 0 {
		return []UploadStatus{}
	}

	var buffer bytes.Buffer
	gob.NewEncoder(&buffer).Encode(&uploads)

	req, err := x.newRequestGZIP(http.MethodPost, x.ServerAddress+"/upload", &buffer)
	if err != nil {
		return uploadStatusFatal(uploads, err.Error())
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", MediaTypeUploadsGob)
	req.Header.Set("Accept", MediaTypeUploadStatusesGob)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return uploadStatusFatal(uploads, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return uploadStatusFatal(uploads, fmt.Sprintf("http request received non-200 status: `%d: %s`", resp.StatusCode, bytes.TrimSpace(body)))

	}

	var results []UploadStatus
	if err := gob.NewDecoder(resp.Body).Decode(&results); err != nil {
		return uploadStatusFatal(uploads, fmt.Sprintf("gob.Decode() failed: %v", err))
	}
	return results
}

func uploadStatusFatal(uploads []Upload, err string) []UploadStatus {
	statuses := make([]UploadStatus, 0, len(uploads))
	for _, upload := range uploads {
		statuses = append(statuses, UploadStatus{
			FileName: upload.FileName,
			Code:     0,
			Error:    err,
		})
	}
	return statuses
}

// Authenticate queries the api server to check that the client has the uploader role.
// An error returns if the query fails or if the client does not have the uploader role.
func (x *Client) Authenticate() error {
	// Query the server
	req, err := x.newRequest(http.MethodGet, x.ServerAddress+"/roles", nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest() failed: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do() failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-200 status `%d: %s`", resp.StatusCode, bytes.TrimSpace(buf))
	}
	// Check that we have the uploader role.
	var roles []login.Role
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return fmt.Errorf("json.Decode() failed: %w", err)
	}
	for _, role := range roles {
		if role == login.RoleVVGOUploader {
			return nil
		}
	}
	return fmt.Errorf("client does not have upload permissions")
}

func (x *Client) newRequestGZIP(method, url string, body io.Reader) (*http.Request, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := io.Copy(gzipWriter, body); err != nil {
		return nil, fmt.Errorf("gzipWriter.Write failed(): %v", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzipWriter.Close() failed(): %v", err)
	}

	req, err := x.newRequest(method, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("gzipWriter.Close() failed(): %v", err)
	}
	req.Header.Set("Content-Encoding", "application/gzip")
	return req, nil
}

func (x *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ClientUserAgent)
	req.Header.Set("Authorization", "Bearer "+x.Token)
	req.Header.Set("Accepts", MediaTypeUploadStatusesGob)
	return req, nil
}

type AsyncClient struct {
	*Client
	AsyncClientConfig
	queue  chan Upload
	status chan UploadStatus
}

type AsyncClientConfig struct {
	MaxParallel int
	QueueLength int
	ClientConfig
}

func NewAsyncClient(conf AsyncClientConfig) *AsyncClient {
	queue := make(chan Upload, conf.QueueLength)
	status := make(chan UploadStatus, conf.QueueLength)
	asyncClient := AsyncClient{
		Client: NewClient(conf.ClientConfig),
		queue:  queue,
		status: status,
	}

	go func() {
		defer close(status)
		for upload := range queue {
			for _, results := range asyncClient.Client.Upload(upload) {
				status <- results
			}
		}
	}()

	return &asyncClient
}

func (x *AsyncClient) Upload(uploads ...Upload) {
	for _, upload := range uploads {
		x.queue <- upload
	}
}

func (x *AsyncClient) Status() <-chan UploadStatus {
	return x.status
}

func (x *AsyncClient) Close() {
	close(x.queue)
}

func (x *AsyncClient) doUploads() {

}
