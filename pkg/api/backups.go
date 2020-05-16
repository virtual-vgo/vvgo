package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type BackupHandler struct {
	Database *Database
	Backups  *storage.Bucket
	NavBar
}

func (x *BackupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "backups_admin_view")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		x.renderView(w, r, ctx)
	case http.MethodPost:
		x.doAction(w, r, ctx)
	}
}

func (x *BackupHandler) renderView(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	type tableRow struct {
		Timestamp    string `json:"timestamp"`
		Size         int64  `json:"size"`
		Version      string `json:"version"`
		DownloadLink string `json:"download_link"`
		RestoreLink  string `json:"restore_link"`
	}

	info := x.Backups.ListObjects(ctx, "")
	rows := make([]tableRow, len(info))
	for i := range info {
		rows[i] = tableRow{
			Timestamp:    info[i].LastModified.String(),
			Size:         info[i].Size,
			DownloadLink: fmt.Sprintf("/download?bucket=%s&object=%s", x.Backups.Name, info[i].Key),
			RestoreLink:  fmt.Sprintf("/backup?cmd=restore&object=%s", info[i].Key),
		}
		rows[i].Size = info[i].Size
		rows[i].Timestamp = info[i].LastModified.Local().Format(time.RFC822)
		fmt.Printf("%#v\n", info[i])
	}

	navBarOpts := x.NavBar.NewOpts(ctx, r)
	page := struct {
		Header    template.HTML
		NavBar    template.HTML
		Rows      []tableRow
		StartLink string
	}{
		Header: Header(),
		NavBar: x.NavBar.RenderHTML(navBarOpts),
		Rows:   rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "backups.gohtml")); !ok {
		internalServerError(w)
		return
	}
	buffer.WriteTo(w)
}

func (x *BackupHandler) doAction(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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

func (x *BackupHandler) restoreFromBucket(ctx context.Context, objectName string) error {
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
		Tags: map[string]string{
			"Vvgo-Api-Version": version.String(),
		},
	}
	name := "dump-" + backup.Timestamp.Format(time.RFC3339) + ".json"
	if err := x.Backups.PutObject(ctx, name, &obj); err != nil {
		return fmt.Errorf("bucket.PutObject() failed: %w", err)
	}
	return nil
}
