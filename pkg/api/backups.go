package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"time"
)

type BackupDocument struct {
	Timestamp time.Time       `json:"timestamp"`
	Parts     []parts.Part    `json:"parts"`
	Version   json.RawMessage `json:"version"`
}

type BackupHandler struct {
	Database *Database
}

func (x *BackupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backup_admin_handler")
	defer span.Send()

	switch r.FormValue("cmd") {
	case "backup":
		if err := x.backupToBucket(ctx); err != nil {
			logger.WithError(err).Error("backup failed")
			internalServerError(w)
			return
		}
	case "restore":
		key := r.FormValue("object")
		if key == "" {
			badRequest(w, "missing form field `object`")
			return
		}
		if err := x.restoreFromBucket(ctx, r.FormValue("object")); err != nil {
			logger.WithError(err).Error("restore failed")
			internalServerError(w)
			return
		}
	default:
		badRequest(w, "missing form field `cmd`")
		return
	}

	if acceptsType(r, "text/html") {
		http.Redirect(w, r, "/backups", http.StatusFound)
	}
}

func (x *BackupHandler) restoreFromBucket(ctx context.Context, objectName string) error {
	var obj storage.Object
	if err := x.Database.Backups.GetObject(ctx, objectName, &obj); err != nil {
		return fmt.Errorf("backups.GetObject() failed: %w", err)
	}

	var document BackupDocument
	if err := json.NewDecoder(bytes.NewReader(obj.Bytes)).Decode(&document); err != nil {
		return fmt.Errorf("json.Decode() failed: %w", err)
	}
	if err := x.Database.Restore(ctx, document); err != nil {
		return fmt.Errorf("database.Restore() failed: %w", err)
	}
	return nil
}

func (x *BackupHandler) backupToBucket(ctx context.Context) error {
	backup, err := x.Database.Backup(ctx)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(backup); err != nil {
		return fmt.Errorf("json.Encode() failed: %w", err)
	}

	obj := storage.Object{
		Bytes:       buf.Bytes(),
		ContentType: "application/json",
	}
	name := "dump-" + backup.Timestamp.Format(time.RFC3339) + ".json"
	if err := x.Database.Backups.PutObject(ctx, name, &obj); err != nil {
		return fmt.Errorf("bucket.PutObject() failed: %w", err)
	}
	return nil
}

func (x *Database) Backup(ctx context.Context) (BackupDocument, error) {
	gotParts, err := x.Parts.List(ctx)
	if err != nil {
		return BackupDocument{}, fmt.Errorf("parts.List() failed: %w", err)
	}
	return BackupDocument{
		Timestamp: time.Now(),
		Parts:     gotParts,
		Version:   version.JSON(),
	}, nil
}

func (x *Database) Restore(ctx context.Context, src BackupDocument) error {
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
