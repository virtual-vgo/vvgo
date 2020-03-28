package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"log"
	"net/http"
)

type ObjectStore interface {
	PutObject(bucketName string, object *Object) error
	ListObjects(bucketName string) []Object
}

type Object struct {
	ContentType string       `json:"content-type"`
	Name        string       `json:"name"`
	Meta        http.Header  `json:"meta"`
	Buffer      bytes.Buffer `json:"-"`
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

type minioDriver struct {
	*minio.Client
}

func NewMinioDriverMust(config MinioConfig) *minioDriver {
	minioClient, err := minio.New(config.Endpoint, config.AccessKeyID,
		config.SecretAccessKey, config.UseSSL)
	if err != nil {
		log.Fatalf("minio.New() failed: %v", err)
	}
	return &minioDriver{minioClient}
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
