package api

import (
	"context"
	"encoding/json"
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
	FileName string
	Code     int
	Error    string
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

func (x *Server) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		invalidContent(w)
		return
	}

	var documents []Upload
	if err := json.NewDecoder(r.Body).Decode(&documents); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		badRequest(w, "")
		return
	}

	if len(documents) == 0 {
		return // nothing to do
	}

	// we'll handle the uploads in goroutines, since these make outgoing http requests to object storage.
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), UploadsTimeout)
	defer cancel()
	wg.Add(len(documents))
	statuses := make(chan UploadStatus, len(documents))
	for _, upload := range documents {
		go func(upload Upload) {
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
		}(upload)
	}

	wg.Wait()
	close(statuses)

	results := make([]UploadStatus, 0, len(documents))
	for status := range statuses {
		results = append(results, status)
	}
	json.NewEncoder(w).Encode(&results)
}

func (x *Server) handleClickTrack(ctx context.Context, upload Upload) UploadStatus {
	panic("Implement me!")

}

func (x *Server) handleSheetMusic(ctx context.Context, upload Upload) UploadStatus {
	panic("Implement me!")
}
