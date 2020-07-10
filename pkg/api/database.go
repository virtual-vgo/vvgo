package api

import (
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/storage"
)

// Database acts as the wrapper/driver for any stateful data.
type Database struct {
	Distro   *storage.Bucket
	Sessions *login.Store
}
