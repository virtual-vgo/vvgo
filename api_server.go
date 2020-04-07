package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"
)

const SheetsBucketName = "sheets"

type ApiServerConfig struct {
	MaxContentLength int64
	BasicAuthUser    string
	BasicAuthPass    string
}

type ApiServer struct {
	ObjectStore
	ApiServerConfig
	*http.ServeMux
	basicAuth
}

func NewApiServer(store ObjectStore, config ApiServerConfig) *ApiServer {
	auth := make(basicAuth)
	if config.BasicAuthUser != "" {
		auth[config.BasicAuthUser] = config.BasicAuthPass
	}
	server := ApiServer{
		ObjectStore:     store,
		ApiServerConfig: config,
		ServeMux:        http.NewServeMux(),
		basicAuth:       auth,
	}

	// debug endpoints from net/http/pprof
	server.HandleFunc("/debug/pprof/", pprof.Index)
	server.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	server.HandleFunc("/debug/pprof/profile", pprof.Profile)
	server.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	server.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server.Handle("/sheets", auth.Authenticate(server.SheetsIndex))
	server.Handle("/sheets/", http.RedirectHandler("/sheets", http.StatusMovedPermanently))
	server.Handle("/sheets/upload", auth.Authenticate(server.SheetsUpload))
	server.Handle("/download", auth.Authenticate(server.Download))
	server.Handle("/version", APIHandlerFunc(server.Version))
	server.Handle("/", http.FileServer(http.Dir("public")))
	return &server
}

type APIHandlerFunc func(w http.ResponseWriter, r *http.Request)

type basicAuth map[string]string

func (x basicAuth) Authenticate(handlerFunc APIHandlerFunc) APIHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			if len(x) == 0 { // skip auth for empty map
				return true
			}
			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(auth) != 2 || auth[0] != "Basic" {
				return false
			}
			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			creds := strings.SplitN(string(payload), ":", 2)
			return len(creds) == 2 && x[creds[0]] == creds[1]
		}(); !ok {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		} else {
			handlerFunc(w, r)
		}
	}
}

// This is http.ResponseWriter middleware that captures the response code
// and other info that might useful in logs
type responseWriter struct {
	code int
	http.ResponseWriter
}

func (x *responseWriter) WriteHeader(code int) {
	x.code = code
	x.ResponseWriter.WriteHeader(code)
}

func (handlerFunc APIHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	results := responseWriter{ResponseWriter: w}
	handlerFunc(&results, r)

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
		"status_code":     results.code,
	}
	switch true {
	case results.code >= 500:
		logger.WithFields(fields).Error("request failed")
	case results.code >= 400:
		logger.WithFields(fields).Error("invalid request")
	default:
		logger.WithFields(fields).Info("request completed")
	}
}

func (x *ApiServer) SheetsIndex(w http.ResponseWriter, r *http.Request) {
	// only accept get
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
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
		if ok := parseAndExecute(w, &rows, "public/sheets.gohtml"); !ok {
			internalServerError(w)
		}
	default:
		if ok := jsonEncode(w, &rows); !ok {
			internalServerError(w)
		}
	}
}

func parseAndExecute(dest io.Writer, data interface{}, templateFiles ...string) bool {
	uploadTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		return false
	}
	if err := uploadTemplate.Execute(dest, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
		return false
	}
	return true
}

func jsonEncode(dest io.Writer, data interface{}) bool {
	if err := json.NewEncoder(dest).Encode(data); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}
	return true
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

func (x *ApiServer) SheetsUpload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if ok := parseAndExecute(w, struct{}{}, "public/sheets/upload.gohtml"); !ok {
			internalServerError(w)
		}

	case http.MethodPost:
		if r.ContentLength > x.MaxContentLength {
			tooManyBytes(w)
			return
		}

		// get the sheet data
		sheet, err := NewSheetFromRequest(r)
		if err != nil {
			logger.WithError(err).Error("NewSheetFromRequest() failed")
			badRequest(w, err.Error())
			return
		}

		// read the pdf content
		var pdfBytes bytes.Buffer
		switch contentType := r.Header.Get("Content-Type"); true {
		case contentType == "application/pdf":
			if _, err = pdfBytes.ReadFrom(r.Body); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				badRequest(w, "")
				return
			}

		case strings.HasPrefix(contentType, "multipart/form-data"):
			file, fileHeader, err := r.FormFile("upload_file")
			if err != nil {
				logger.WithError(err).Error("r.FormFile() failed")
				badRequest(w, "")
				return
			}
			defer file.Close()

			if contentType := fileHeader.Header.Get("Content-Type"); contentType != "application/pdf" {
				logger.WithField("Content-Type", contentType).Error("invalid content type")
				invalidContent(w)
				return
			}

			// read the pdf from the body
			if _, err = pdfBytes.ReadFrom(file); err != nil {
				logger.WithError(err).Error("r.body.Read() failed")
				badRequest(w, "")
				return
			}

		default:
			logger.WithField("Content-Type", contentType).Error("invalid content type")
			invalidContent(w)
			return
		}

		// check file type
		if contentType := http.DetectContentType(pdfBytes.Bytes()); contentType != "application/pdf" {
			logger.WithField("Detected-Content-Type", contentType).Error("invalid content type")
			invalidContent(w)
			return
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
			internalServerError(w)
			return
		}

		// redirect web browsers back to /sheets/upload
		if acceptsType(r, "text/html") {
			http.Redirect(w, r, "/sheets", http.StatusFound)
		}

	default:
		methodNotAllowed(w)
	}
}

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
}

type UploadStatus struct {
	FileName string
	Code     int
	Error    string
}

type ClixUpload struct {
	ContentType string   `json:"content_type"`
	PartNames   []string `json:"part_names"`
	PartNumbers []int    `json:"part_numbers"`
}

type SheetsUpload struct {
	ContentType string   `json:"content_type"`
	PartNames   []string `json:"part_names"`
	PartNumbers []int    `json:"part_numbers"`
}

const UploadsTimeout = 10 * time.Second

func (x *ApiServer) Upload(w http.ResponseWriter, r *http.Request) {
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

func (x *ApiServer) handleClickTrack(ctx context.Context, upload Upload) UploadStatus {
	panic("Implement me!")

}

func (x *ApiServer) handleSheetMusic(ctx context.Context, upload Upload) UploadStatus {
	panic("Implement me!")
}

func (x *ApiServer) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	values := r.URL.Query()
	key := values.Get("key")
	bucket := values.Get("bucket")

	downloadURL, err := x.DownloadURL(bucket, key)
	switch e := err.(type) {
	case nil:
		http.Redirect(w, r, downloadURL, http.StatusFound)
	case minio.ErrorResponse:
		if e.StatusCode == http.StatusNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
	default:
		logger.WithError(err).Error("minio.StatObject() failed")
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (x ApiServer) Version(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	versionHeader := version.Header()
	for k := range versionHeader {
		w.Header().Set(k, versionHeader.Get(k))
	}

	switch true {
	case acceptsType(r, "application/json"):
		w.Write(version.JSON())
	default:
		w.Write([]byte(version.String()))
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
