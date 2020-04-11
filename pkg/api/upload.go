package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clix"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type UploadType string

const (
	UploadTypeClix   UploadType = "clix"
	UploadTypeSheets UploadType = "sheets"
)

type UploadHandler struct{ *Database }

type Upload struct {
	UploadType    `json:"upload_type"`
	*ClixUpload   `json:"clix_upload"`
	*SheetsUpload `json:"sheets_upload"`
	Project       string `json:"project"`
	FileName      string `json:"file_name"`
	FileBytes     []byte `json:"file_bytes"`
	ContentType   string `json:"content_type"`
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
				statuses <- handleClickTrack(ctx, x.Clix, upload)
			case UploadTypeSheets:
				statuses <- handleSheetMusic(ctx, x.Sheets, upload)
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

func handleClickTrack(ctx context.Context, gotClix clix.Clix, upload *Upload) UploadStatus {
	if status := upload.ValidateClix(); status != uploadSuccess(upload) {
		return status
	}
	file := clix.File{
		MediaType: upload.ContentType,
		Ext:       filepath.Ext(upload.FileName),
		Bytes:     upload.FileBytes,
	}
	if ok := gotClix.Store(ctx, upload.Clix(), &file); !ok {
		return uploadInternalServerError(upload)
	}
	return uploadSuccess(upload)
}

func (upload *Upload) ValidateClix() UploadStatus {
	// verify that we have all the necessary info
	clixUpload := upload.ClixUpload
	if clixUpload == nil {
		return uploadBadRequest(upload, "missing json field `sheets_upload`")
	}

	if len(clixUpload.PartNames) == 0 {
		return uploadBadRequest(upload, "missing part names")
	}

	if len(clixUpload.PartNumbers) == 0 {
		return uploadBadRequest(upload, "missing part numbers")
	}

	// verify content type
	if !strings.HasPrefix(upload.ContentType, "audio/") {
		logger.WithField("Content-Type", upload.ContentType).Error("invalid content type")
		return uploadInvalidContent(upload)
	}

	// verify the file contents
	if contentType := http.DetectContentType(upload.FileBytes); !strings.HasPrefix(contentType, "audio/") {
		logger.WithField("Detected-Content-Type", contentType).Error("invalid content type")
		return uploadInvalidContent(upload)
	}
	return uploadSuccess(upload)
}

func (upload *Upload) Clix() []clix.Click {
	clixUpload := upload.ClixUpload
	// convert the upload into sheets
	gotClix := make([]clix.Click, 0, len(clixUpload.PartNames)*len(clixUpload.PartNumbers))
	for _, partName := range clixUpload.PartNames {
		for _, partNumber := range clixUpload.PartNumbers {
			gotClix = append(gotClix, clix.Click{
				Project:    upload.Project,
				PartName:   partName,
				PartNumber: partNumber,
			})
		}
	}
	return gotClix
}

func handleSheetMusic(ctx context.Context, sheets sheets.Sheets, upload *Upload) UploadStatus {
	if status := upload.ValidateSheets(); status != uploadSuccess(upload) {
		return status
	}
	if ok := sheets.Store(ctx, upload.Sheets(), upload.FileBytes); !ok {
		return uploadInternalServerError(upload)
	}
	return uploadSuccess(upload)
}

func (upload *Upload) ValidateSheets() UploadStatus {
	// verify that we have all the necessary info
	sheetsUpload := upload.SheetsUpload
	if sheetsUpload == nil {
		return uploadBadRequest(upload, "missing json field `sheets_upload`")
	}

	if len(sheetsUpload.PartNames) == 0 {
		return uploadBadRequest(upload, "missing part names")
	}

	if len(sheetsUpload.PartNumbers) == 0 {
		return uploadBadRequest(upload, "missing part numbers")
	}

	// verify content type
	if upload.ContentType != "application/pdf" {
		logger.WithField("Content-Type", upload.ContentType).Error("invalid content type")
		return uploadInvalidContent(upload)
	}

	// verify the file contents
	if contentType := http.DetectContentType(upload.FileBytes); contentType != "application/pdf" {
		logger.WithField("Detected-Content-Type", contentType).Error("invalid content type")
		return uploadInvalidContent(upload)
	}
	return uploadSuccess(upload)
}

func (upload *Upload) Sheets() []sheets.Sheet {
	sheetsUpload := upload.SheetsUpload
	// convert the upload into sheets
	gotSheets := make([]sheets.Sheet, 0, len(sheetsUpload.PartNames)*len(sheetsUpload.PartNumbers))
	for _, partName := range sheetsUpload.PartNames {
		for _, partNumber := range sheetsUpload.PartNumbers {
			gotSheets = append(gotSheets, sheets.Sheet{
				Project:    upload.Project,
				PartName:   partName,
				PartNumber: partNumber,
			})
		}
	}
	return gotSheets
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
