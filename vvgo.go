package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v6"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const Location = "us-east-1"

func main() {
	// We will use the MinIO server running at https://play.min.io in this example.
	// Feel free to use this service for testing and development.
	// Access credentials shown in this example are open to the public.
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// build minio client
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("minio.New() failed: %v", err)
	}
	objectStore := minioDriver{minioClient}

	apiServer := ApiServer{objectStore}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", apiServer.Upload)
	mux.HandleFunc("/index", apiServer.Index)
	mux.Handle("/", http.FileServer(http.Dir(".")))
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(httpServer.ListenAndServe())
}

type ObjectStore interface {
	PutObject(bucketName string, object *Object) error
	ListObjects(bucketName string) []Object
}

type minioDriver struct {
	*minio.Client
}

type Object struct {
	ContentType string       `json:"content-type"`
	Name        string       `json:"name"`
	Meta        http.Header  `json:"meta"`
	Buffer      bytes.Buffer `json:"-"`
}

func (x *minioDriver) PutObject(bucketName string, object *Object) error {
	// make the bucket if it doesn't exist
	if err := x.MakeBucket(bucketName); err != nil {
		return err
	}
	n, err := x.Client.PutObject(bucketName, object.Name, &object.Buffer,
		int64(object.Buffer.Len()), minio.PutObjectOptions{ContentType: object.ContentType})
	if err != nil {
		return err
	}
	log.Printf("uploaded %s of size %d\n", object.Name, n)
	return nil
}

func (x *minioDriver) MakeBucket(bucketName string) error {
	exists, err := x.BucketExists(bucketName)
	if err != nil {
		return fmt.Errorf("x.minioClient.BucketExists() failed: %v", err)
	}
	if exists == false {
		if err := x.Client.MakeBucket(bucketName, Location); err != nil {
			return fmt.Errorf("x.minioClient.MakeBucket() failed: %v", err)
		}
	}
	return nil
}

func (x *minioDriver) ListObjects(bucketName string) []Object {
	done := make(chan struct{})
	defer close(done)

	var objects []Object
	for objectInfo := range x.Client.ListObjects(bucketName, "", false, done) {
		objects = append(objects, Object{
			ContentType: objectInfo.ContentType,
			Name:        objectInfo.Key,
			Meta:        objectInfo.Metadata,
		})
	}
	return objects
}

type ApiServer struct {
	minioDriver
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
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// encode the metadata
	var metaBytes bytes.Buffer
	if err := json.NewEncoder(&metaBytes).Encode(&meta); err != nil {
		log.Printf("json.Encode() failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// write the metadata
	metaObject := Object{
		ContentType: "application/json",
		Name:        fmt.Sprintf("%s_%s_%d_meta.json", meta.Project, meta.Instrument, meta.PartNumber),
		Buffer:      metaBytes,
	}
	if err := x.PutObject("music_pdfs", &metaObject); err != nil {
		log.Printf("storage.PutObject() failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		Name:        fmt.Sprintf("%s_%s_%d.pdf", meta.Project, meta.Instrument, meta.PartNumber),
		Buffer:      pdfBytes,
	}
	if err := x.PutObject("music_pdfs", &object); err != nil {
		log.Printf("storage.PutObject() failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	header.Add("project", x.Project)
	header.Add("instrument", x.Instrument)
	header.Add("part_number", strconv.Itoa(x.PartNumber))
	return header
}

func (x *MusicPDFMeta) ReadFromHeader(header http.Header) {
	x.Project = header.Get("project")
	x.Instrument = header.Get("instrument")
	x.PartNumber, _ = strconv.Atoi(header.Get("part_number"))
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
