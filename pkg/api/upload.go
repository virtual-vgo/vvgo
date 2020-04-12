package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type UploadHandler struct{ *Database }

type UploadType string

const (
	UploadTypeClix   UploadType = "clix"
	UploadTypeSheets UploadType = "sheets"
)

type Upload struct {
	UploadType    `json:"upload_type"`
	*ClixUpload   `json:"clix_upload"`
	*SheetsUpload `json:"sheets_upload"`
	Project       string `json:"project"`
	FileName      string `json:"file_name"`
	FileBytes     []byte `json:"file_bytes"`
	ContentType   string `json:"content_type"`
}

var (
	ErrMissingClix        = fmt.Errorf("missing field `clix_upload`")
	ErrMissingSheets      = fmt.Errorf("missing field `sheets_upload`")
	ErrMissingPartNames   = fmt.Errorf("missing part names")
	ErrMissingPartNumbers = fmt.Errorf("missing part numbers")
)

func (upload *Upload) ValidateClix() error {
	// verify that we have all the necessary info
	clixUpload := upload.ClixUpload
	if clixUpload == nil {
		return ErrMissingClix
	} else if len(clixUpload.PartNames) == 0 {
		return ErrMissingPartNames
	} else if len(clixUpload.PartNumbers) == 0 {
		return ErrMissingPartNumbers
	} else {
		return nil
	}
}

func (upload *Upload) ValidateSheets() error {
	// verify that we have all the necessary info
	sheetsUpload := upload.SheetsUpload
	if sheetsUpload == nil {
		return ErrMissingSheets
	} else if len(sheetsUpload.PartNames) == 0 {
		return ErrMissingPartNames
	} else if len(sheetsUpload.PartNumbers) == 0 {
		return ErrMissingPartNumbers
	} else {
		return nil
	}
}

func (upload *Upload) File() *storage.File {
	return &storage.File{
		MediaType: upload.ContentType,
		Ext:       filepath.Ext(upload.FileName),
		Bytes:     upload.FileBytes,
	}
}

type UploadStatus struct {
	FileName string `json:"file_name"`
	Code     int    `json:"code"`
	Error    string `json:"error,omitempty"`
}

type ClixUpload struct {
	PartNames   []string `json:"part_names"`
	PartNumbers []uint8  `json:"part_numbers"`
}

type SheetsUpload struct {
	PartNames   []string `json:"part_names"`
	PartNumbers []uint8  `json:"part_numbers"`
}

const UploadsTimeout = 10 * time.Second

var ErrInvalidMediaType = fmt.Errorf("invalid media type")

func (x UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		invalidContent(w)
		return
	}

	var documents []Upload
	if ok := jsonDecode(r.Body, &documents); !ok {
		badRequest(w, "")
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

			// check for context cancelled
			select {
			case <-ctx.Done():
				statuses <- UploadStatus{
					FileName: upload.FileName,
					Code:     http.StatusRequestTimeout,
					Error:    ctx.Err().Error(),
				}
			default:
			}

			// check that the project exists
			if !projects.Exists(upload.Project) {
				statuses <- UploadStatus{
					FileName: upload.FileName,
					Code:     http.StatusBadRequest,
					Error:    "project not found",
				}
			}

			// handle the upload
			switch upload.UploadType {
			case UploadTypeClix:
				statuses <- x.handleClickTrack(ctx, upload)
			case UploadTypeSheets:
				statuses <- x.handleSheetMusic(ctx, upload)
			default:
				statuses <- UploadStatus{
					FileName: upload.FileName,
					Code:     http.StatusBadRequest,
					Error:    "invalid upload type",
				}
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
	if err := upload.ValidateClix(); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	file := upload.File()
	if err := file.ValidateMediaType("audio/"); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	objectKey, ok := x.Clix.PutFile(file)
	if !ok {
		return uploadInternalServerError(upload)
	}

	// update parts with the revision
	gotParts := makeParts(upload.Project, upload.ClixUpload.PartNames, upload.ClixUpload.PartNumbers)
	for i := range gotParts {
		gotParts[i].Click = objectKey
	}
	if ok := x.Parts.Save(ctx, gotParts); !ok {
		return uploadInternalServerError(upload)
	} else {
		return uploadSuccess(upload)
	}
}

func (x UploadHandler) handleSheetMusic(ctx context.Context, upload *Upload) UploadStatus {
	if err := upload.ValidateSheets(); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	// upload the click track
	file := storage.File{
		MediaType: upload.ContentType,
		Ext:       filepath.Ext(upload.FileName),
		Bytes:     upload.FileBytes,
	}
	if err := file.ValidateMediaType("audio/"); err != nil {
		return uploadBadRequest(upload, err.Error())
	}

	objectKey, ok := x.Clix.PutFile(&file)
	if !ok {
		return uploadInternalServerError(upload)
	}

	gotParts := makeParts(upload.Project, upload.ClixUpload.PartNames, upload.ClixUpload.PartNumbers)
	for i := range gotParts {
		gotParts[i].Click = objectKey
	}
	if ok := x.Parts.Save(ctx, gotParts); !ok {
		return uploadInternalServerError(upload)
	} else {
		return uploadSuccess(upload)
	}
}

func (upload *Upload) RenderParts() []parts.Part {
	var names []string
	var numbers []uint8
	switch upload.UploadType {
	case UploadTypeClix:
		names = upload.ClixUpload.PartNames
		numbers = upload.ClixUpload.PartNumbers
	case UploadTypeSheets:
		names = upload.SheetsUpload.PartNames
		numbers = upload.SheetsUpload.PartNumbers
	}
	return makeParts(upload.Project, names, numbers)
}

func makeParts(project string, names []string, numbers []uint8) []parts.Part {
	allParts := make([]parts.Part, 0, len(names)*len(numbers))
	for _, name := range names {
		for _, number := range numbers {
			allParts = append(allParts, parts.Part{
				ID: parts.ID{Project: project, Name: name, Number: number},
			})
		}
	}
	return allParts
}

func uploadSuccess(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusOK,
	}
}

func uploadNotImplemented(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusNotImplemented,
		Error:    "lo sentimos, no implementado",
	}
}

func uploadBadRequest(upload *Upload, reason string) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusBadRequest,
		Error:    reason,
	}
}

func uploadInvalidContent(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusUnsupportedMediaType,
		Error:    "unsupported media type",
	}
}

func uploadInternalServerError(upload *Upload) UploadStatus {
	return UploadStatus{
		FileName: upload.FileName,
		Code:     http.StatusInternalServerError,
		Error:    "internal server error",
	}
}
