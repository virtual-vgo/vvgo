package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"golang.org/x/net/html"
	"os"
)

func init() {
	PublicFiles = "../../public"
}

type MockBucket struct {
	putObject   func(name string, object *storage.Object) bool
	getObject   func(name string, dest *storage.Object) bool
	downloadURL func(name string) (string, error)
}

func (x *MockBucket) PutObject(name string, object *storage.Object) bool {
	return x.putObject(name, object)
}

func (x *MockBucket) GetObject(name string, dest *storage.Object) bool {
	return x.getObject(name, dest)
}

func (x *MockBucket) DownloadURL(name string) (string, error) {
	return x.downloadURL(name)
}

func tokenizeHTMLFile(src string) *html.Tokenizer {
	file, err := os.Open(src)
	if err != nil {
		panic(fmt.Errorf("os.Open() failed: %v", err))
	}
	return html.NewTokenizer(file)
}
