package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"path/filepath"
	"strings"
)

func (x *Server) SheetsIndex(w http.ResponseWriter, r *http.Request) {
	// only accept get
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	type tableRow struct {
		sheets.Sheet
		Link string `json:"link"`
	}

	objects := x.ListObjects(sheets.BucketName)
	rows := make([]tableRow, 0, len(objects))
	for i := range objects {
		sheet := sheets.NewSheetFromTags(objects[i].Tags)
		rows = append(rows, tableRow{
			Sheet: sheet,
			Link:  sheet.Link(),
		})
	}

	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(w, &rows, filepath.Join(Public, "sheets.gohtml")); !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}
	default:
		if ok := jsonEncode(w, &rows); !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}

func (x *Server) SheetsUpload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if ok := parseAndExecute(w, struct{}{}, filepath.Join(Public, "sheets", "upload.gohtml")); !ok {
			http.Error(w, "", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if r.ContentLength > x.MaxContentLength {
			http.Error(w, "", http.StatusRequestEntityTooLarge)
			return
		}

		// get the sheet data
		sheet, err := sheets.NewSheetFromRequest(r)
		if err != nil {
			logger.WithError(err).Error("NewSheetFromRequest() failed")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// read the pdf content
		var pdfBytes bytes.Buffer
		switch contentType := r.Header.Get("Content-Type"); true {
		case contentType == "application/pdf":
			if _, err = pdfBytes.ReadFrom(r.Body); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				http.Error(w, "", http.StatusBadRequest)
				return
			}

		case strings.HasPrefix(contentType, "multipart/form-data"):
			file, fileHeader, err := r.FormFile("upload_file")
			if err != nil {
				logger.WithError(err).Error("r.FormFile() failed")
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			defer file.Close()

			if contentType := fileHeader.Header.Get("Content-Type"); contentType != "application/pdf" {
				logger.WithField("Content-Type", contentType).Error("invalid content type")
				http.Error(w, "", http.StatusUnsupportedMediaType)
				return
			}

			// read the pdf from the body
			if _, err = pdfBytes.ReadFrom(file); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				http.Error(w, "", http.StatusBadRequest)
				return
			}

		default:
			logger.WithField("Content-Type", contentType).Error("invalid content type")
			http.Error(w, "", http.StatusUnsupportedMediaType)
			return
		}

		// check file type
		if contentType := http.DetectContentType(pdfBytes.Bytes()); contentType != "application/pdf" {
			logger.WithField("Detected-Content-Type", contentType).Error("invalid content type")
			http.Error(w, "", http.StatusUnsupportedMediaType)
			return
		}

		// write the pdf
		object := storage.Object{
			ContentType: "application/pdf",
			Name:        sheet.ObjectKey(),
			Tags:        sheet.Tags(),
			Buffer:      pdfBytes,
		}
		if err := x.PutObject(sheets.BucketName, &object); err != nil {
			logger.WithError(err).Error("storage.PutObject() failed")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// redirect web browsers back to /sheets/upload
		if acceptsType(r, "text/html") {
			http.Redirect(w, r, "/sheets", http.StatusFound)
		}

	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}
