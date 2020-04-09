package api

import (
	"bytes"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
)

func init() {
	Public = "../../public"
}

var _ storage.ObjectStorage = new(MockObjectStore)

type MockObjectStore struct {
	putObject   func(bucketName string, object *storage.Object) error
	listObjects func(bucketName string) []storage.Object
	downloadURL func(bucketName string, objectName string) (string, error)
}

func (x MockObjectStore) PutObject(bucketName string, object *storage.Object) error {
	return x.putObject(bucketName, object)
}

func (x MockObjectStore) ListObjects(bucketName string) []storage.Object {
	return x.listObjects(bucketName)
}

func (x MockObjectStore) DownloadURL(bucketName string, objectName string) (string, error) {
	return x.downloadURL(bucketName, objectName)
}

func tokenizeHTMLFile(src string) *html.Tokenizer {
	file, err := os.Open(src)
	if err != nil {
		panic(fmt.Errorf("os.Open() failed: %v", err))
	}
	return html.NewTokenizer(file)
}

func newPostRequest(uri string, contentType string, src io.Reader) *http.Request {
	req, err := http.NewRequest(http.MethodPost, uri, src)
	if err != nil {
		panic(fmt.Sprintf("http.NewRequest() failed: %v", err))
	}
	req.Header.Set("Content-Type", contentType)
	return req
}

// Creates a new file upload http request with optional extra params
func newFileUploadRequest(uri string, params map[string]string, fileParam, fileName, contentType string, src io.Reader) *http.Request {
	escapeQuotes := strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace

	if r, err := func() (*http.Request, error) {
		var body bytes.Buffer
		multipartWriter := multipart.NewWriter(&body)

		fileHeader := make(textproto.MIMEHeader)
		fileHeader.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fileParam), escapeQuotes(fileName)))
		fileHeader.Set("Content-Type", contentType)
		fileDest, err := multipartWriter.CreatePart(fileHeader)

		if err != nil {
			return nil, fmt.Errorf("multipartWriter.CreateFormFile() failed: %v", err)
		}

		if _, err = io.Copy(fileDest, src); err != nil {
			return nil, fmt.Errorf("io.Copy() failed: %v", err)
		}

		for key, val := range params {
			if err = multipartWriter.WriteField(key, val); err != nil {
				return nil, fmt.Errorf("multipartWriter.WriteField() failed: %v", err)
			}
		}

		if err = multipartWriter.Close(); err != nil {
			return nil, fmt.Errorf("multipartWriter.Close() failed: %v", err)
		}

		if req, err := http.NewRequest("POST", uri, &body); err != nil {
			return nil, fmt.Errorf("http.NewRequest() failed: %v", err)
		} else {
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
			return req, nil
		}
	}(); err != nil {
		panic(err)
	} else {
		return r
	}
}

func mustReadFile(fileName string) []byte {
	if buf, err := ioutil.ReadFile(fileName); err != nil {
		panic(fmt.Sprintf("ioutil.ReadFile() failed: %v", err))
	} else {
		return buf
	}
}
