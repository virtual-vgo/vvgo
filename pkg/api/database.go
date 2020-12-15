package api

import (
	"github.com/virtual-vgo/vvgo/pkg/login"
)

// Database acts as the wrapper/driver for any stateful data.
type Database struct {
	Sessions      *login.Store
}
