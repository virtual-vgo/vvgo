package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SheetsBucketName = "sheets"

type ApiServerConfig struct {
	MaxContentLength int64
}

type ApiServer struct {
	ObjectStore
	ApiServerConfig
	*http.ServeMux
}

func NewApiServer(store ObjectStore, config ApiServerConfig) *ApiServer {
	server := ApiServer{
		ObjectStore:     store,
		ApiServerConfig: config,
		ServeMux:        http.NewServeMux(),
	}
	server.Handle("/sheets", APIHandlerFunc(server.SheetsIndex))
	server.Handle("/sheets/upload", APIHandlerFunc(server.SheetsUpload))
	server.Handle("/download", http.HandlerFunc(server.Download))
	server.Handle("/", http.FileServer(http.Dir("public")))
	return &server
}

type APIHandlerFunc func(r *http.Request) ([]byte, int)

func (x APIHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	defer r.Body.Close()
	body, code := logRequest(x, r)
	w.WriteHeader(code)
	w.Write(body)
}

func logRequest(handlerFunc APIHandlerFunc, r *http.Request) ([]byte, int) {
	start := time.Now()
	body, code := handlerFunc(r)

	clientIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	fields := logrus.Fields{
		"client_ip":       clientIP,
		"request_path":    r.URL.EscapedPath(),
		"user_agent":      r.UserAgent(),
		"request_method":  r.Method,
		"request_size":    r.ContentLength,
		"request_seconds": time.Since(start).Seconds(),
		"status_code":     code,
	}
	switch true {
	case code >= 500:
		logger.WithFields(fields).Error("request failed")
	case 400 <= code && code < 500:
		logger.WithFields(fields).Error("invalid request")
	default:
		logger.WithFields(fields).Info("request completed")
	}
	return body, code
}

func (x *ApiServer) SheetsIndex(r *http.Request) ([]byte, int) {
	// only accept get
	if r.Method != http.MethodGet {
		return nil, http.StatusMethodNotAllowed
	}

	type tableRow struct {
		Sheet
		Link string `json:"link"`
	}

	objects := x.ListObjects(SheetsBucketName)
	rows := make([]tableRow, 0, len(objects))
	for i := range objects {
		sheet := NewSheetFromTags(objects[i].Tags)
		rows = append(rows, tableRow{
			Sheet: sheet,
			Link:  sheet.Link(),
		})
	}

	switch true {
	case acceptsType(r, "text/html"):
		return parseAndExecute(&rows, "public/sheets.gohtml")
	default:
		return jsonEncode(&rows)
	}
}

func jsonEncode(data interface{}) ([]byte, int) {
	if dataJSON, err := json.Marshal(data); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return nil, http.StatusInternalServerError
	} else {
		return dataJSON, http.StatusOK
	}
}

func acceptsType(r *http.Request, mimeType string) bool {
	for _, value := range r.Header["Accept"] {
		for _, wantType := range strings.Split(value, ",") {
			if wantType == mimeType {
				return true
			}
		}
	}
	return false
}

func (x *ApiServer) SheetsUpload(r *http.Request) ([]byte, int) {
	switch r.Method {
	case http.MethodGet:
		return parseAndExecute(struct{}{}, "public/sheets/upload.gohtml")

	case http.MethodPost:
		if r.ContentLength > x.MaxContentLength {
			return nil, http.StatusRequestEntityTooLarge
		}

		// get the sheet data
		sheet, err := NewSheetFromRequest(r)
		if err != nil {
			logger.WithError(err).Error("NewSheetFromRequest() failed")
			return []byte(err.Error()), http.StatusBadRequest
		}

		// read the pdf content
		var pdfBytes bytes.Buffer
		switch contentType := r.Header.Get("Content-Type"); true {
		case contentType == "application/pdf":
			if _, err = pdfBytes.ReadFrom(r.Body); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				return nil, http.StatusBadRequest
			}

		case strings.HasPrefix(contentType, "multipart/form-data"):
			file, fileHeader, err := r.FormFile("upload_file")
			if err != nil {
				logger.WithError(err).Error("r.FormFile() failed")
				return nil, http.StatusBadRequest
			}
			defer file.Close()

			if contentType := fileHeader.Header.Get("Content-Type"); contentType != "application/pdf" {
				logger.WithField("Content-Type", contentType).Error("invalid content type")
				return nil, http.StatusUnsupportedMediaType
			}

			// read the pdf from the body
			if _, err = pdfBytes.ReadFrom(file); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				return nil, http.StatusBadRequest
			}

		default:
			logger.WithField("Content-Type", contentType).Error("invalid content type")
			return nil, http.StatusUnsupportedMediaType
		}

		// check file type
		if contentType := http.DetectContentType(pdfBytes.Bytes()); contentType != "application/pdf" {
			logger.WithField("Detected-Content-Type", contentType).Error("invalid content type")
			return nil, http.StatusUnsupportedMediaType
		}

		// write the pdf
		object := Object{
			ContentType: "application/pdf",
			Name:        sheet.ObjectKey(),
			Tags:        sheet.Tags(),
			Buffer:      pdfBytes,
		}
		if err := x.PutObject(SheetsBucketName, &object); err != nil {
			logger.WithError(err).Error("storage.PutObject() failed")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusOK
	default:
		return nil, http.StatusMethodNotAllowed
	}
}

func parseAndExecute(data interface{}, templateFiles ...string) ([]byte, int) {
	var buf bytes.Buffer
	uploadTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		return nil, http.StatusInternalServerError
	}
	if err := uploadTemplate.Execute(&buf, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
		return nil, http.StatusInternalServerError
	}
	return buf.Bytes(), http.StatusOK
}

func (x *ApiServer) Download(w http.ResponseWriter, r *http.Request) {
	var downloadURL string
	body, code := logRequest(func(*http.Request) ([]byte, int) {
		if r.Method != http.MethodGet {
			return nil, http.StatusMethodNotAllowed
		}

		values := r.URL.Query()
		key := values.Get("key")
		bucket := values.Get("bucket")

		var err error
		downloadURL, err = x.DownloadURL(bucket, key)

		switch e := err.(type) {
		case nil:
			return nil, http.StatusFound
		case minio.ErrorResponse:
			if e.StatusCode == http.StatusNotFound {
				return nil, http.StatusNotFound
			}
		}

		logger.WithError(err).Error("minio.StatObject() failed")
		return nil, http.StatusInternalServerError
	}, r)

	switch code {
	case http.StatusFound:
		http.Redirect(w, r, downloadURL, code)
	case http.StatusNotFound:
		http.NotFound(w, r)
	default:
		w.WriteHeader(code)
		w.Write(body)
	}
}

var (
	ErrMissingProject    = fmt.Errorf("missing required field `project`")
	ErrMissingInstrument = fmt.Errorf("missing required field `instrument`")
	ErrMissingPartNumber = fmt.Errorf("missing required field `part_number`")
)

type Sheet struct {
	Project    string `json:"project"`
	Instrument string `json:"instrument"`
	PartNumber int    `json:"part_number"`
}

func NewSheetFromTags(tags Tags) Sheet {
	partNumber, _ := strconv.Atoi(tags["Part-Number"])
	return Sheet{
		Project:    tags["Project"],
		Instrument: tags["Instrument"],
		PartNumber: partNumber,
	}
}

func NewSheetFromRequest(r *http.Request) (Sheet, error) {
	partNumber, _ := strconv.Atoi(r.FormValue("part_number"))
	sheet := Sheet{
		Project:    r.FormValue("project"),
		Instrument: r.FormValue("instrument"),
		PartNumber: partNumber,
	}
	return sheet, sheet.Validate()
}

func (x Sheet) ObjectKey() string {
	return fmt.Sprintf("%s-%s-%d.pdf", x.Project, x.Instrument, x.PartNumber)
}

func (x Sheet) Link() string {
	return fmt.Sprintf("/download?bucket=%s&key=%s", SheetsBucketName, x.ObjectKey())
}

func (x Sheet) Tags() map[string]string {
	return map[string]string{
		"Project":     x.Project,
		"Instrument":  x.Instrument,
		"Part-Number": strconv.Itoa(x.PartNumber),
	}
}

func (x Sheet) Validate() error {
	if x.Project == "" {
		return ErrMissingProject
	} else if x.Instrument == "" {
		return ErrMissingInstrument
	} else if x.PartNumber == 0 {
		return ErrMissingPartNumber
	} else {
		return nil
	}
}
