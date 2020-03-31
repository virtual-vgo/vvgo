package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
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

	server.Handle("/sheets", auth.Authenticate(server.SheetsIndex))
	server.Handle("/sheets/upload", auth.Authenticate(server.SheetsUpload))
	server.Handle("/download", auth.Authenticate(server.Download))
	server.Handle("/", http.FileServer(http.Dir("public")))
	return &server
}

type APIHandlerFunc func(w http.ResponseWriter, r *http.Request)

type basicAuth map[string]string

func (x basicAuth) Authenticate(handlerFunc APIHandlerFunc) APIHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(x) > 0 { // skip auth for empty map
			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(auth) != 2 || auth[0] != "Basic" {
				http.Error(w, "authorization failed", http.StatusUnauthorized)
				return
			}
			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			creds := strings.SplitN(string(payload), ":", 2)
			if len(creds) != 2 || !(x[creds[0]] == creds[1]) {
				http.Error(w, "authorization failed", http.StatusUnauthorized)
				return
			}
		}
		handlerFunc(w, r)
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
		http.Error(w, "", http.StatusMethodNotAllowed)
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
			http.Error(w, "", http.StatusInternalServerError)
		}
	default:
		if ok := jsonEncode(w, &rows); !ok {
			http.Error(w, "", http.StatusInternalServerError)
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
			http.Error(w, "", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if r.ContentLength > x.MaxContentLength {
			http.Error(w, "", http.StatusRequestEntityTooLarge)
			return
		}

		// get the sheet data
		sheet, err := NewSheetFromRequest(r)
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
		object := Object{
			ContentType: "application/pdf",
			Name:        sheet.ObjectKey(),
			Tags:        sheet.Tags(),
			Buffer:      pdfBytes,
		}
		if err := x.PutObject(SheetsBucketName, &object); err != nil {
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

func (x *ApiServer) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
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
