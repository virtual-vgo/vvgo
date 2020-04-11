package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/sheet"
	"net/http"
	"sync"
	"time"
)

type UploadType string

func (x UploadType) String() string { return string(x) }

const (
	UploadTypeClix   UploadType = "clix"
	UploadTypeSheets UploadType = "sheets"
)

type UploadHandler struct {
	sheet.Sheets
}

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
	Error    string `json:"error"`
}

type ClixUpload struct {
	PartNames   []string `json:"part_names"`
	PartNumbers []int    `json:"part_numbers"`
}

type SheetsUpload struct {
	PartNames   []string `json:"part_names"`
	PartNumbers []int    `json:"part_numbers"`
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
			if !projectExists(upload.Project) {
				statuses <- UploadStatus{
					FileName: upload.FileName,
					Code:     http.StatusBadRequest,
					Error:    "project not found",
				}
			}

			// handle the upload
			switch upload.UploadType {
			case UploadTypeClix:
				statuses <- handleClickTrack(ctx, upload)
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

func handleClickTrack(_ context.Context, upload *Upload) UploadStatus {
	return uploadNotImplemented(upload)
}

func handleSheetMusic(ctx context.Context, sheets sheet.Sheets, upload *Upload) UploadStatus {
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

func (upload *Upload) Sheets() []sheet.Sheet {
	sheetsUpload := upload.SheetsUpload
	// convert the upload into sheets
	gotSheets := make([]sheet.Sheet, 0, len(sheetsUpload.PartNames)*len(sheetsUpload.PartNumbers))
	for _, partName := range sheetsUpload.PartNames {
		for _, partNumber := range sheetsUpload.PartNumbers {
			gotSheets = append(gotSheets, sheet.Sheet{
				Project:    upload.Project,
				PartName:   partName,
				PartNumber: partNumber,
			})
		}
	}
	return gotSheets
}

func projectExists(project string) bool {
	return project == "01-snake-eater"
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
