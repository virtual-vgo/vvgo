package api

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const UploadsTimeout = 5 * 60 * time.Second

type UploadHandler struct{ *Database }

type UploadType string

func (x UploadType) String() string { return string(x) }

const (
	UploadTypeClix   UploadType = "clix"
	UploadTypeSheets UploadType = "sheets"
)

const MediaTypeUploadsGob = "application/x.vvgo.pkg.api.uploads.gob"

type Uploads []Upload

type Upload struct {
	UploadType  `json:"upload_type"`
	PartNames   []string `json:"part_names"`
	Project     string   `json:"project"`
	FileName    string   `json:"file_name"`
	FileBytes   []byte   `json:"file_bytes"`
	ContentType string   `json:"content_type"`
}

var (
	ErrInvalidUploadType = errors.New("invalid upload type")
	ErrMissingProject    = errors.New("missing project")
	ErrMissingPartNames  = errors.New("missing part names")
	ErrEmptyFileBytes    = errors.New("empty file bytes")
)

func (upload *Upload) Validate() error {
	switch {
	case upload.Project == "":
		return ErrMissingProject
	case projects.Exists(upload.Project) == false:
		return projects.ErrNotFound
	case len(upload.PartNames) == 0:
		return ErrMissingPartNames
	case parts.ValidNames(upload.PartNames...) == false:
		return parts.ErrInvalidPartName
	case len(upload.FileBytes) == 0:
		return ErrEmptyFileBytes
	}

	file := upload.File()
	switch upload.UploadType {
	case UploadTypeClix:
		return file.ValidateMediaType("audio/")
	case UploadTypeSheets:
		return file.ValidateMediaType("application/pdf")
	default:
		return ErrInvalidUploadType
	}
}

func (upload *Upload) File() *storage.File {
	return &storage.File{
		MediaType: upload.ContentType,
		Ext:       filepath.Ext(upload.FileName),
		Bytes:     upload.FileBytes,
	}
}

func (upload *Upload) Parts() []parts.Part {
	allParts := make([]parts.Part, 0, len(upload.PartNames))
	for _, name := range upload.PartNames {
		allParts = append(allParts, parts.Part{
			ID: parts.ID{
				Project: strings.TrimSpace(strings.ToLower(upload.Project)),
				Name:    strings.TrimSpace(strings.ToLower(name)),
			},
		})
	}
	return allParts
}

const MediaTypeUploadStatusesGob = "application/x.vvgo.pkg.api.upload_statuses.gob"

type UploadStatus struct {
	FileName string `json:"file_name"`
	Code     int    `json:"code"`
	Error    string `json:"error,omitempty"`
}

func (x UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "upload_handler")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var uploads Uploads
	if ok := readUploads(r, &uploads); !ok {
		badRequest(w, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, UploadsTimeout)
	defer cancel()
	statuses := x.ServeUploads(ctx, uploads)

	switch {
	case acceptsType(r, MediaTypeUploadStatusesGob):
		if err := gob.NewEncoder(w).Encode(&statuses); err != nil {
			logger.WithError(err).Error("gob.Encode() failed", err)
		}

	default:
		if err := json.NewEncoder(w).Encode(&statuses); err != nil {
			logger.WithError(err).Error("json.Encode() failed", err)
		}
	}
}

func readUploads(r *http.Request, dest *Uploads) bool {
	var body bytes.Buffer
	if ok := readBody(&body, r); !ok {
		return false
	}

	switch r.Header.Get("Content-Type") {
	case MediaTypeUploadsGob:
		return gobDecode(&body, dest)
	case "application/json":
		return jsonDecode(&body, dest)
	default:
		return false
	}
}

func (x *UploadHandler) ServeUploads(ctx context.Context, uploads Uploads) []UploadStatus {
	// we'll handle the uploads in goroutines, since these make outgoing http requests to object storage.
	var wg sync.WaitGroup
	wg.Add(len(uploads))
	statusesCh := make(chan UploadStatus, len(uploads))
	for _, upload := range uploads {
		go func(upload Upload) {
			defer wg.Done()
			statusesCh <- x.serveUpload(ctx, upload)
		}(upload)
	}
	wg.Wait()
	close(statusesCh)

	statuses := make([]UploadStatus, 0, len(uploads))
	for status := range statusesCh {
		statuses = append(statuses, status)
	}
	return statuses
}

func (x UploadHandler) serveUpload(ctx context.Context, upload Upload) UploadStatus {
	// validate the upload
	if err := upload.Validate(); err != nil {
		return uploadBadRequest(&upload, err.Error())
	}

	// check if context cancelled
	select {
	case <-ctx.Done():
		return uploadCtxCancelled(&upload, ctx.Err())
	default:
	}

	// handle the upload
	switch upload.UploadType {
	case UploadTypeClix:
		return x.handleClix(ctx, &upload)
	case UploadTypeSheets:
		return x.handleSheets(ctx, &upload)
	default:
		return uploadBadRequest(&upload, "invalid upload type")
	}
}

func (x UploadHandler) handleClix(ctx context.Context, upload *Upload) UploadStatus {
	file := upload.File()
	if err := file.ValidateMediaType("audio/"); err != nil {
		logger.WithError(err).Error("file.ValidateMediaType() failed")
		return uploadBadRequest(upload, err.Error())
	}

	if err := x.Distro.PutFile(ctx, file); err != nil {
		logger.WithError(err).Error("x.Clix.PutFile() failed")
		return uploadInternalServerError(upload)
	}
	return x.handleParts(ctx, upload, file.ObjectKey())
}

func (x UploadHandler) handleSheets(ctx context.Context, upload *Upload) UploadStatus {
	file := upload.File()
	if err := file.ValidateMediaType("application/pdf"); err != nil {
		logger.WithError(err).Error("file.ValidateMediaType() failed")
		return uploadBadRequest(upload, err.Error())
	}

	if err := x.Distro.PutFile(ctx, file); err != nil {
		logger.WithError(err).Error("x.Sheets.PutFile() failed")
		return uploadInternalServerError(upload)
	} else {
		return x.handleParts(ctx, upload, file.ObjectKey())
	}
}

func (x UploadHandler) handleParts(ctx context.Context, upload *Upload, objectKey string) UploadStatus {
	// update parts with the revision
	uploadParts := upload.Parts()
	for i := range uploadParts {
		switch upload.UploadType {
		case UploadTypeSheets:
			uploadParts[i].Sheets.NewKey(objectKey)
		case UploadTypeClix:
			uploadParts[i].Clix.NewKey(objectKey)
		}
	}
	if err := x.Parts.Save(ctx, uploadParts); err != nil {
		logger.WithError(err).Error("x.RedisParts.Save() failed")
		return uploadInternalServerError(upload)
	} else {
		return uploadSuccess(upload)
	}
}

func uploadSuccess(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusOK,
	}
}

func uploadCtxCancelled(upload *Upload, err error) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusRequestTimeout,
		Error:    err.Error(),
	}
}

func uploadBadRequest(upload *Upload, reason string) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusBadRequest,
		Error:    reason,
	}
}

func uploadInternalServerError(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusInternalServerError,
		Error:    "internal server error",
	}
}
