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
	"net/url"
	"strconv"
	"strings"
	"time"
)

const MusicPdfsBucketName = "sheets"

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

	objects := x.ListObjects(MusicPdfsBucketName)
	pdfs := make([]MusicPDFMeta, 0, len(objects))
	for i := range objects {
		pdfs = append(pdfs, NewMusicPDFMetaFromTags(objects[i].Tags))
	}

	var buf bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		musicPDFsTemplate, err := template.ParseFiles("public/music_pdfs.gohtml")
		if err != nil {
			logger.WithError(err).Error("template.ParseFiles() failed")
			return nil, http.StatusInternalServerError
		}
		if err := musicPDFsTemplate.Execute(&buf, &pdfs); err != nil {
			logger.WithError(err).Error("template.Execute() failed")
			return nil, http.StatusInternalServerError
		}
	default:
		if err := json.NewEncoder(&buf).Encode(&objects); err != nil {
			logger.WithError(err).Error("json.Encode() failed")
			return nil, http.StatusInternalServerError
		}
	}
	return buf.Bytes(), http.StatusOK
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
	// only accept post
	if r.Method != http.MethodPost {
		return nil, http.StatusMethodNotAllowed
	}

	if r.ContentLength > x.MaxContentLength {
		return nil, http.StatusRequestEntityTooLarge
	}

	// read the metadata
	meta := NewMusicPDFMetaFromUrlValues(r.URL.Query())
	if err := meta.Validate(); err != nil {
		return []byte(err.Error()), http.StatusBadRequest
	}

	// validate content encoding
	if r.Header.Get("Content-Type") != "application/pdf" {
		return nil, http.StatusUnsupportedMediaType
	}

	// read the pdf from the body
	var pdfBytes bytes.Buffer
	if _, err := pdfBytes.ReadFrom(r.Body); err != nil {
		logger.WithError(err).Error("r.body.Read() failed")
		return nil, http.StatusBadRequest
	}

	// write the pdf
	object := Object{
		ContentType: "application/pdf",
		Name:        fmt.Sprintf("%s-%s-%d.pdf", meta.Project, meta.Instrument, meta.PartNumber),
		Tags:        meta.ToTags(),
		Buffer:      pdfBytes,
	}
	if err := x.PutObject(MusicPdfsBucketName, &object); err != nil {
		logger.WithError(err).Error("storage.PutObject() failed")
		return nil, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

const LinkExpiration = 24 * 3600 * time.Second // 1 Day

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

type MusicPDFMeta struct {
	Project      string
	Instrument   string
	PartNumber   int
	DownloadLink string
}

func NewMusicPDFMetaFromTags(tags Tags) MusicPDFMeta {
	partNumber, _ := strconv.Atoi(tags["Part-Number"])
	return MusicPDFMeta{
		Project:    tags["Project"],
		Instrument: tags["Instrument"],
		PartNumber: partNumber,
	}
}

func NewMusicPDFMetaFromUrlValues(values url.Values) MusicPDFMeta {
	partNumber, _ := strconv.Atoi(values.Get("part_number"))
	return MusicPDFMeta{
		Project:    values.Get("project"),
		Instrument: values.Get("instrument"),
		PartNumber: partNumber,
	}
}

func (x *MusicPDFMeta) ToTags() map[string]string {
	return map[string]string{
		"Project":     x.Project,
		"Instrument":  x.Instrument,
		"Part-Number": strconv.Itoa(x.PartNumber),
	}
}

func (x *MusicPDFMeta) Validate() error {
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
