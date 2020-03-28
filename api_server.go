package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type ApiServer struct {
	ObjectStore
}

func (x *ApiServer) Index(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// only accept get
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	objects := x.ListObjects("music_pdfs")
	if err := json.NewEncoder(w).Encode(&objects); err != nil {
		log.Printf("json.Encode() failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (x *ApiServer) Upload(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// only accept post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// only allow <1MB
	if r.ContentLength > int64(1e6) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// read the metadata
	var meta MusicPDFMeta
	meta.ReadFromUrlValues(r.URL.Query())
	if err := meta.Validate(); err != nil {
		log.Printf("validation failed for: %#v", meta)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// read the pdf from the body
	var pdfBytes bytes.Buffer
	if _, err := pdfBytes.ReadFrom(r.Body); err != nil {
		log.Printf("r.body.Read() failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write the pdf
	object := Object{
		ContentType: "application/pdf",
		Name:        fmt.Sprintf("%s-%s-%d.pdf", meta.Project, meta.Instrument, meta.PartNumber),
		Meta:        meta.ToHeader(),
		Buffer:      pdfBytes,
	}
	if err := x.PutObject("music_pdfs", &object); err != nil {
		log.Printf("storage.PutObject() failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

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

func (x *MusicPDFMeta) ToHeader() http.Header {
	header := make(http.Header)
	header.Add("Project", x.Project)
	header.Add("Instrument", x.Instrument)
	header.Add("PartNumber", strconv.Itoa(x.PartNumber))
	return header
}

func (x *MusicPDFMeta) ReadFromHeader(header http.Header) {
	x.Project = header.Get("Project")
	x.Instrument = header.Get("Instrument")
	x.PartNumber, _ = strconv.Atoi(header.Get("PartNumber"))
}

func (x *MusicPDFMeta) ReadFromUrlValues(values url.Values) {
	x.Project = values.Get("project")
	x.Instrument = values.Get("instrument")
	x.PartNumber, _ = strconv.Atoi(values.Get("part_number"))
}

func (x *MusicPDFMeta) Validate() error {
	if x.Project == "" {
		return ErrMissingProject
	}
	if x.Instrument == "" {
		return ErrMissingInstrument
	}
	if x.PartNumber == 0 {
		return ErrMissingPartNumber
	}
	{
		return nil
	}
}
