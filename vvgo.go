package main

import (
	"github.com/minio/minio-go/v6"
	"log"
	"net/http"
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
	apiServer := ApiServer{&minioDriver{minioClient}}

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
