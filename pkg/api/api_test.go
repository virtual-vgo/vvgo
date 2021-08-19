package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
)

func init() { PublicFiles = "../../public" }

func backgroundContext() context.Context {
	return sheets.NoOpSheets(context.Background())
}
