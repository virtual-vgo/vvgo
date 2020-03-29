package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

const Location = "us-east-1"

var logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	},
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.InfoLevel,
	ExitFunc:     os.Exit,
	ReportCaller: false,
}

type Config struct {
	Minio MinioConfig
	Api   ApiServerConfig
}

func NewDefaultConfig() Config {
	return Config{
		Minio: MinioConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			UseSSL:          false,
		},
		Api: ApiServerConfig{
			MaxContentLength: 1e6,
		},
	}
}

func (x *Config) ParseEnv() {
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		x.Minio.Endpoint = endpoint
	}
	if id := os.Getenv("MINIO_ACCESS_KEY_ID"); id != "" {
		x.Minio.AccessKeyID = id
	}
	if key := os.Getenv("MINIO_SECRET_ACCESS_KEY"); key != "" {
		x.Minio.SecretAccessKey = key
	}
	x.Minio.UseSSL, _ = strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	if maxContentLength, _ := strconv.ParseInt(os.Getenv("API_MAX_CONTENT_LENGTH"), 10, 64); maxContentLength != 0 {
		x.Api.MaxContentLength = maxContentLength
	}
}

func main() {
	config := NewDefaultConfig()
	config.ParseEnv()

	apiServer := NewApiServer(
		NewMinioDriverMust(config.Minio),
		config.Api,
	)

	mux := http.NewServeMux()
	mux.Handle("/music_pdfs", APIHandlerFunc(apiServer.MusicPDFsIndex))
	mux.Handle("/music_pdfs/upload", APIHandlerFunc(apiServer.MusicPDFsUpload))
	mux.Handle("/", http.FileServer(http.Dir("public")))
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	logger.Fatal(httpServer.ListenAndServe())
}
