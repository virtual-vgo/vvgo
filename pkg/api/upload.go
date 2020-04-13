package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type UploadHandler struct{ *Storage }

type UploadType string

func (x UploadType) String() string { return string(x) }

const (
	UploadTypeClix   UploadType = "clix"
	UploadTypeSheets UploadType = "sheets"
)

type Upload struct {
	UploadType  `json:"upload_type"`
	PartNames   []string `json:"part_names"`
	PartNumbers []uint8  `json:"part_numbers"`
	Project     string   `json:"project"`
	FileName    string   `json:"file_name"`
	FileBytes   []byte   `json:"file_bytes"`
	ContentType string   `json:"content_type"`
}

var (
	ErrInvalidUploadType  = errors.New("invalid upload type")
	ErrMissingProject     = errors.New("missing project")
	ErrMissingPartNames   = errors.New("missing part names")
	ErrMissingPartNumbers = errors.New("missing part numbers")
	ErrEmptyFileBytes     = errors.New("empty file bytes")
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
	case len(upload.PartNumbers) == 0:
		return ErrMissingPartNumbers
	case parts.ValidNumbers(upload.PartNumbers...) == false:
		return parts.ErrInvalidPartNumber
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
	allParts := make([]parts.Part, 0, len(upload.PartNames)*len(upload.PartNumbers))
	for _, name := range upload.PartNames {
		for _, number := range upload.PartNumbers {
			allParts = append(allParts, parts.Part{
				ID: parts.ID{
					Project: strings.TrimSpace(strings.ToLower(upload.Project)),
					Name:    strings.TrimSpace(strings.ToLower(name)),
					Number:  number,
				},
			})
		}
	}
	return allParts
}

type UploadStatus struct {
	FileName string `json:"file_name"`
	Code     int    `json:"code"`
	Error    string `json:"error,omitempty"`
}

const UploadsTimeout = 5 * 60 * time.Second

func (x UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var documents []Upload
	switch r.Header.Get("Content-Type") {
	case "application/octet-stream":
		if ok := gobDecode(r.Body, &documents); !ok {
			badRequest(w, "")
			return
		}

	case "application/json":
		if ok := jsonDecode(r.Body, &documents); !ok {
			badRequest(w, "")
			return
		}

	default:
		invalidContent(w)
		return
	}

	// we'll handle the uploads in goroutines, since these make outgoing http requests to object storage.
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), UploadsTimeout)
	defer cancel()
	wg.Add(len(documents))
	statuses := make(chan UploadStatus, len(documents))
	for _, upload := range documents {
		go func(upload *Upload) {
			defer wg.Done()

			// validate the upload
			if err := upload.Validate(); err != nil {
				statuses <- uploadBadRequest(upload, err.Error())
				return
			}

			// check for context cancelled
			select {
			case <-ctx.Done():
				statuses <- uploadCtxCancelled(upload, ctx.Err())
			default:
			}

			// handle the upload
			switch upload.UploadType {
			case UploadTypeClix:
				statuses <- x.handleClickTrack(ctx, upload)
			case UploadTypeSheets:
				statuses <- x.handleSheetMusic(ctx, upload)
			}
		}(&upload)
	}

	wg.Wait()
	close(statuses)

	results := make([]UploadStatus, 0, len(documents))
	for status := range statuses {
		results = append(results, status)
	}
	json.NewEncoder(w).Encode(&results)
}

func (x UploadHandler) handleClickTrack(ctx context.Context, upload *Upload) UploadStatus {
	file := upload.File()
	if err := file.ValidateMediaType("audio/"); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	if ok := x.Clix.PutFile(file); !ok {
		return uploadInternalServerError(upload)
	} else {
		return x.handleParts(ctx, upload, file.ObjectKey())
	}
}

func (x UploadHandler) handleSheetMusic(ctx context.Context, upload *Upload) UploadStatus {
	file := upload.File()
	if err := file.ValidateMediaType("application/pdf"); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	if ok := x.Sheets.PutFile(file); !ok {
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
	if ok := x.Parts.Save(ctx, uploadParts); !ok {
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
