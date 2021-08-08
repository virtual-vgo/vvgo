package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
)

func init() {
	PublicFiles = "../../public"
	redis.InitializeFromEnv()
}

func backgroundContext() context.Context {
	return sheets.NoOpSheets(context.Background())
}
