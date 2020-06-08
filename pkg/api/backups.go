package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type BackupHandler struct {
	Database *Database
	Backups  *storage.Bucket
}

func (x BackupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_admin_view")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		x.renderView(w, r, ctx)
	case http.MethodPost:
		x.doAction(w, r, ctx)
	}
}

func (x BackupHandler) renderView(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	type tableRow struct {
		Timestamp    string `json:"timestamp"`
		SizeKB       int64  `json:"size_kb"`
		Version      string `json:"version"`
		Object       string `json:"object"`
		DownloadLink string `json:"download_link"`
	}

	info := x.Backups.ListObjects(ctx, "backups/")
	rows := make([]tableRow, len(info))
	for i := range info {
		dlValues := make(url.Values)
		dlValues.Add("bucket", x.Backups.Name)
		dlValues.Add("object", info[i].Key)
		rows[i] = tableRow{
			Timestamp:    info[i].LastModified.Local().Format(time.RFC822),
			SizeKB:       info[i].Size / 1000,
			DownloadLink: "/download?" + dlValues.Encode(),
			Object:       info[i].Key,
		}
	}

	opts := NewNavBarOpts(ctx)
	opts.BackupsActive = true
	page := struct {
		NavBar NavBarOpts
		Rows   []tableRow
	}{
		NavBar: opts,
		Rows:   rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "backups.gohtml")); !ok {
		internalServerError(w)
		return
	}
	buffer.WriteTo(w)
}

func (x BackupHandler) doAction(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}
	return
}

func (x BackupHandler) restoreFromBucket(ctx context.Context, objectName string) error {
	var obj storage.Object
	if err := x.Backups.GetObject(ctx, objectName, &obj); err != nil {
		return fmt.Errorf("backups.GetObject() failed: %w", err)
	}

	var document DatabaseBackup
	if err := json.NewDecoder(bytes.NewReader(obj.Bytes)).Decode(&document); err != nil {
		return fmt.Errorf("json.Decode() failed: %w", err)
	}
	if err := x.Database.Restore(ctx, document); err != nil {
		return fmt.Errorf("database.Restore() failed: %w", err)
	}
	return nil
}

func (x BackupHandler) backupToBucket(ctx context.Context) error {
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
		Tags: map[string]string{
			"Vvgo-Api-Version": version.String(),
		},
	}
	name := "backups/dump-" + backup.Timestamp.Format(time.RFC3339) + ".json"
	if err := x.Backups.PutObject(ctx, name, &obj); err != nil {
		return fmt.Errorf("bucket.PutObject() failed: %w", err)
	}
	return nil
}
