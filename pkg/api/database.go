package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"time"
)

type Database struct {
	Parts  *parts.RedisParts
	Distro *storage.Bucket
}

type DatabaseBackup struct {
	Timestamp  time.Time       `json:"timestamp"`
	Parts      []parts.Part    `json:"parts"`
	ApiVersion json.RawMessage `json:"api_version"`
}

func (x *Database) Backup(ctx context.Context) (DatabaseBackup, error) {
	gotParts, err := x.Parts.List(ctx)
	if err != nil {
		return DatabaseBackup{}, fmt.Errorf("parts.List() failed: %w", err)
	}
	return DatabaseBackup{
		Timestamp:  time.Now(),
		Parts:      gotParts,
		ApiVersion: version.JSON(),
	}, nil
}

func (x *Database) Restore(ctx context.Context, src DatabaseBackup) error {
	// truncate existing dbs
	if err := x.Parts.DeleteAll(ctx); err != nil {
		return fmt.Errorf("parts.DeleteAll() failed: %w", err)
	}
	// save the new parts
	if err := x.Parts.Save(ctx, src.Parts); err != nil {
		return fmt.Errorf("parts.Save() failed: %w", err)
	}
	return nil
}
