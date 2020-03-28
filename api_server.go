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
		"user_agent":      r.UserAgent(),
		"request_method":  r.Method,
		"request_size":    r.ContentLength,
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
	pdfs := make([]MusicPDFMeta, 0, len(objects))
	for i := range objects {
		pdfs = append(pdfs, NewMusicPDFMetaFromTags(objects[i].Tags))
	}

	var buf bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		musicPDFsTemplate, err := template.ParseFiles("public/music_pdfs.gohtml")
		if err != nil {
			logger.Printf("template.ParseFiles() failed: %v", err)
			return nil, http.StatusInternalServerError
		}
		if err := musicPDFsTemplate.Execute(&buf, &pdfs); err != nil {
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

func acceptsType(r *http.Request, mimeType string) bool {
	for _, t := range strings.Split(r.Header.Get("Accept"), ",") {
		if t == mimeType {
			return true
		}
	}
	return false
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
	meta := NewMusicPDFMetaFromUrlValues(r.URL.Query())
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
		Tags:        meta.ToTags(),
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
