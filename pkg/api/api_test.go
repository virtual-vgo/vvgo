package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/storage"
)

func init() {
	PublicFiles = "../../public"
}

type MockBucket struct {
	putFile     func(ctx context.Context, file *storage.File) bool
	downloadURL func(ctx context.Context, name string) (string, error)
	getObject   func(ctx context.Context, name string, dest *storage.Object) bool
	putObject   func(ctx context.Context, name string, object *storage.Object) bool
}

func (x *MockBucket) PutFile(ctx context.Context, file *storage.File) bool {
	return x.putFile(ctx, file)
}

func (x *MockBucket) DownloadURL(ctx context.Context, name string) (string, error) {
	return x.downloadURL(ctx, name)
}

func (x *MockBucket) GetObject(ctx context.Context, name string, dest *storage.Object) bool {
	return x.getObject(ctx, name, dest)
}

func (x *MockBucket) PutObject(ctx context.Context, name string, object *storage.Object) bool {
	return x.putObject(ctx, name, object)
}

type MockLocker struct {
	lock   func(ctx context.Context) bool
	unlock func(ctx context.Context)
}

func (x *MockLocker) Lock(ctx context.Context) bool {
	return x.lock(ctx)
}

func (x *MockLocker) Unlock(ctx context.Context) {
	x.unlock(ctx)
}
