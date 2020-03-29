package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
)

type ObjectStore interface {
	PutObject(bucketName string, object *Object) error
	ListObjects(bucketName string) []Object
}

type Object struct {
	ContentType string       `json:"content-type"`
	Name        string       `json:"name"`
	Tags        Tags         `json:"tags"`
	Buffer      bytes.Buffer `json:"-"`
}

type Tags map[string]string

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
		logger.WithError(err).Fatalf("minio.New() failed")
	}
	return &minioDriver{minioClient}
}

func (x *minioDriver) PutObject(bucketName string, object *Object) error {
	// make the bucket if it doesn't exist
	if err := x.MakeBucket(bucketName); err != nil {
		return err
	}

	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.Client.PutObject(bucketName, object.Name, &object.Buffer, -1, opts)
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"object_name": object.Name,
		"object_size": n,
	}).Info("uploaded pdf")
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
		if objectInfo.Key == "" {
			continue
		}

		opts := minio.StatObjectOptions{}
		objectInfo, err := x.Client.StatObject(bucketName, objectInfo.Key, opts)
		if err != nil {
			logger.WithError(err).Error("minio.StatObject() failed")
			continue
		}

		objects = append(objects, Object{
			ContentType: objectInfo.ContentType,
			Name:        objectInfo.Key,
			Tags:        Tags(objectInfo.UserMetadata),
		})
	}
	return objects
}
