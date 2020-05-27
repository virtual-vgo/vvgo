package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"time"
)

// Database acts as the wrapper/driver for any stateful data.
type Database struct {
	Parts    *parts.RedisParts
	Distro   *storage.Bucket
	Sessions *login.Store
}

// DatabaseBackup is a document containing a snapshot of the stateful data we want to backup.
type DatabaseBackup struct {
	// Timestamp when the backup document was created.
	Timestamp time.Time `json:"timestamp"`

	// Parts is the slice of _all_ parts in the database.
	Parts []parts.Part `json:"parts"`

	// ApiVersion is the version of the api server when this database backup was created.
	ApiVersion json.RawMessage `json:"api_version"`
}

// Backup creates a new DatabaseBackup document.
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

// Restore restores the database from the given document.
// This is a destructive operation, and the first step is to truncate existing Parts data.
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
