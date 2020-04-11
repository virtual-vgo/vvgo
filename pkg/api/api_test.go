package api

import (
	"github.com/virtual-vgo/vvgo/pkg/storage"
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
