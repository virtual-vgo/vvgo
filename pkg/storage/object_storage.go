package storage

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"net/url"
	"time"
)

var logger = log.Logger()

type ObjectStorage interface {
	PutObject(bucketName string, object *Object) error
	ListObjects(bucketName string) []Object
	DownloadURL(bucketName string, objectName string) (string, error)
}

type Object struct {
	ContentType string
	Name        string
	Tags        Tags
	Buffer      bytes.Buffer
}

type Tags map[string]string

type MinioConfig struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type minioDriver struct {
	MinioConfig
	*minio.Client
}

func NewMinioDriverMust(config MinioConfig) *minioDriver {
	minioClient, err := minio.New(config.Endpoint, config.AccessKey,
		config.SecretKey, config.UseSSL)
	if err != nil {
		logger.WithError(err).Fatalf("minio.New() failed")
	}
	return &minioDriver{
		MinioConfig: config,
		Client:      minioClient,
	}
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
		if err := x.Client.MakeBucket(bucketName, x.Region); err != nil {
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

		object, err := x.StatObject(bucketName, objectInfo.Key)
		if err != nil {
			logger.WithError(err).Error("minio.StatObject() failed")
			continue
		}

		objects = append(objects, object)
	}
	return objects
}

func (x *minioDriver) StatObject(bucketName, objectName string) (Object, error) {
	opts := minio.StatObjectOptions{}
	objectInfo, err := x.Client.StatObject(bucketName, objectName, opts)
	if err != nil {
		return Object{}, err
	} else {
		return Object{
			ContentType: objectInfo.ContentType,
			Name:        objectInfo.Key,
			Tags:        Tags(objectInfo.UserMetadata),
		}, nil
	}
}

const LinkExpiration = 24 * 3600 * time.Second // 1 Day for protect links

func (x *minioDriver) DownloadURL(bucketName string, objectName string) (string, error) {
	policy, err := x.Client.GetBucketPolicy(bucketName)
	if err != nil {
		return "", err
	}

	var downloadUrl *url.URL
	switch policy {
	case "download":
		downloadUrl, err = url.Parse(fmt.Sprintf("%s/%s/%s", x.Client.EndpointURL(), bucketName, objectName))
	default:
		downloadUrl, err = x.Client.PresignedGetObject(bucketName, objectName, LinkExpiration, nil)
	}
	if err != nil {
		return "", err
	} else {
		return downloadUrl.String(), nil
	}
}
