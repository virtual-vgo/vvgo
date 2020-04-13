package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/storage"
)

func init() {
	PublicFiles = "../../public"
}

type MockBucket struct {
	putFile     func(file *storage.File) bool
	downloadURL func(name string) (string, error)
	getObject   func(name string, dest *storage.Object) bool
	putObject   func(name string, object *storage.Object) bool
}

func (x *MockBucket) PutFile(file *storage.File) bool {
	return x.putFile(file)
}

func (x *MockBucket) DownloadURL(name string) (string, error) {
	return x.downloadURL(name)
}

func (x *MockBucket) GetObject(name string, dest *storage.Object) bool {
	return x.getObject(name, dest)
}

func (x *MockBucket) PutObject(name string, object *storage.Object) bool {
	return x.putObject(name, object)
}

type MockLocker struct {
	lock   func(ctx context.Context) bool
	unlock func()
}

func (x *MockLocker) Lock(ctx context.Context) bool {
	return x.lock(ctx)
}

func (x *MockLocker) Unlock() {
	x.unlock()
}
