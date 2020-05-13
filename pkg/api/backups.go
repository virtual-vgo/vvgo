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
	Storage
}

func (x BackupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_handler")
	defer span.Send()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	document, err := x.Backup(ctx)
	if err != nil {
		logger.WithError(err).Error("Backup() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &document)
}

type RestoreHandler struct {
	Storage
}

func (x RestoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_handler")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	// read the request body
	var buf bytes.Buffer
	if ok := readBody(&buf, r); !ok {
		badRequest(w, "invalid body")
		return
	}

	// read the document
	var document BackupDocument
	if err := json.NewDecoder(&buf).Decode(&document); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		badRequest(w, err.Error())
		return
	}

	if err := x.Restore(ctx, document); err != nil {
		logger.WithError(err).Error("Restore() failed")
		internalServerError(w)
		return
	}
}

func (x *Storage) Backup(ctx context.Context) (BackupDocument, error) {
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

func (x *Storage) Restore(ctx context.Context, src BackupDocument) error {
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

func BackupToBucket(ctx context.Context, db *Storage, bucket *storage.Bucket) error {
	// make the backup
	backup, err := db.Backup(ctx)
	if err != nil {
		return err
	}

	backupBytes, err := json.Marshal(&backup)
	if err != nil {
		return fmt.Errorf("json.Marshal() failed: %w", err)
	}

	obj := storage.Object{
		ContentType: "application/json",
		Tags: map[string]string{
			"api_version": version.String(),
		},
		Bytes: backupBytes,
	}
	name := "dump." + backup.Timestamp.Format(time.RFC3339) + ".json.gz"
	return bucket.PutObject(ctx, name, &obj)
}
