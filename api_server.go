package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ApiServer struct {
	ObjectStore
	mux *http.ServeMux
}

var musicPDFsTemplate = template.Must(template.ParseFiles("templates/music_pdfs.gohtml"))

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
	logger.WithFields(logrus.Fields{
		"client_ip":       clientIP,
		"request_path":    r.URL.EscapedPath(),
		"request_method":  r.Method,
		"request_size":   r.ContentLength,
		"request_seconds": time.Since(start).Seconds(),
		"status_code":     code,
	}).Info("request completed")
	return body, code
}

func (x *ApiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if x.mux == nil {
		x.mux = http.NewServeMux()
		x.mux.Handle("/music_pdfs", APIHandlerFunc(x.MusicPDFsIndex))
		x.mux.Handle("/music_pdfs/upload", APIHandlerFunc(x.MusicPDFsUpload))
	}
	x.mux.ServeHTTP(w, r)
}

func (x *ApiServer) MusicPDFsIndex(r *http.Request) ([]byte, int) {
	// only accept get
	if r.Method != http.MethodGet {
		return nil, http.StatusMethodNotAllowed
	}

	objects := x.ListObjects(MusicPdfsBucketName)
	var buf bytes.Buffer
	switch r.Header.Get("Accept") {
	case "text/html":
		if err := musicPDFsTemplate.Execute(&buf, &objects); err != nil {
			logger.Printf("template.Execute() failed: %v", err)
			return nil, http.StatusInternalServerError
		}
	default:
		if err := json.NewEncoder(&buf).Encode(&objects); err != nil {
			logger.Printf("json.Encode() failed: %v", err)
			return nil, http.StatusInternalServerError
		}
	}
	return buf.Bytes(), http.StatusOK
}

func (x *ApiServer) MusicPDFsUpload(r *http.Request) ([]byte, int) {
	// only accept post
	if r.Method != http.MethodPost {
		return nil, http.StatusMethodNotAllowed
	}

	// only allow <1MB
	if r.ContentLength > int64(1e6) {
		return nil, http.StatusRequestEntityTooLarge
	}

	// read the metadata
	var meta MusicPDFMeta
	meta.ReadFromUrlValues(r.URL.Query())
	if err := meta.Validate(); err != nil {
		return []byte(err.Error()), http.StatusBadRequest
	}

	// read the pdf from the body
	var pdfBytes bytes.Buffer
	if _, err := pdfBytes.ReadFrom(r.Body); err != nil {
		logger.WithError(err).Println("r.body.Read() failed")
		return nil, http.StatusBadRequest
	}

	// write the pdf
	object := Object{
		ContentType: "application/pdf",
		Name:        fmt.Sprintf("%s-%s-%d.pdf", meta.Project, meta.Instrument, meta.PartNumber),
		Meta:        meta.ToMap(),
		Buffer:      pdfBytes,
	}
	if err := x.PutObject(MusicPdfsBucketName, &object); err != nil {
		logger.WithError(err).Println("storage.PutObject() failed")
		return nil, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

const MusicPdfsBucketName = "music-pdfs"

var (
	ErrMissingProject    = fmt.Errorf("missing required field `project`")
	ErrMissingInstrument = fmt.Errorf("missing required field `instrument`")
	ErrMissingPartNumber = fmt.Errorf("missing required field `part_number`")
)

type MusicPDFMeta struct {
	Project    string
	Instrument string
	PartNumber int
}

func (x *MusicPDFMeta) ToMap() map[string]string {
	return map[string]string{
		"Project":     x.Project,
		"Instrument":  x.Instrument,
		"Part-Number": strconv.Itoa(x.PartNumber),
	}
}

func (x *MusicPDFMeta) ReadFromHeader(header http.Header) {
	x.Project = header.Get("Project")
	x.Instrument = header.Get("Instrument")
	x.PartNumber, _ = strconv.Atoi(header.Get("Part-Number"))
}

func (x *MusicPDFMeta) ReadFromUrlValues(values url.Values) {
	x.Project = values.Get("project")
	x.Instrument = values.Get("instrument")
	x.PartNumber, _ = strconv.Atoi(values.Get("part_number"))
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
