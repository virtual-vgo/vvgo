package api

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const ClientUserAgent = "Virtual-VGO client"

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
