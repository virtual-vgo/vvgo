package storage

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

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

type MinioDriver struct {
	MinioConfig
	*minio.Client
}

func NewMinioDriverMust(config MinioConfig) *MinioDriver {
	minioClient, err := minio.New(config.Endpoint, config.AccessKey,
		config.SecretKey, config.UseSSL)
	if err != nil {
		logger.WithError(err).Fatalf("minio.New() failed")
	}
	return &MinioDriver{
		MinioConfig: config,
		Client:      minioClient,
	}
}

func (x *MinioDriver) PutObject(bucketName string, object *Object) bool {
	// make the bucket if it doesn't exist
	if ok := x.MakeBucket(bucketName); !ok {
		return false
	}

	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.Client.PutObject(bucketName, object.Name, &object.Buffer, -1, opts)
	if err != nil {
		logger.WithError(err).Error("minioClient.PutObject() failed")
		return false
	}
	logger.WithFields(logrus.Fields{
		"object_name": object.Name,
		"object_size": n,
	}).Info("uploaded object")
	return true
}

func (x *MinioDriver) GetObject(bucketName string, objectName string, dest *Object) bool {
	minioObject, err := x.Client.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		logger.WithError(err).Error("minioClient.GetObject() failed")
		return false
	}
	info, err := minioObject.Stat()
	if err != nil {
		logger.WithError(err).Error("minioObject.Stat() failed")
		return false
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(minioObject); err != nil {
		logger.WithError(err).Error("minioObject.Read() failed")
		return false
	}
	*dest = Object{
		ContentType: info.ContentType,
		Name:        info.Key,
		Tags:        map[string]string(info.UserMetadata),
		Buffer:      buffer,
	}
	return true
}

func (x *MinioDriver) MakeBucket(bucketName string) bool {
	exists, err := x.BucketExists(bucketName)
	if err != nil {
		logger.WithError(err).Error("minioClient.BucketExists() failed")
		return false
	}
	if exists == false {
		if err := x.Client.MakeBucket(bucketName, x.Region); err != nil {
			logger.WithError(err).Error("minioClient.MakeBucket() failed")
			return false
		}
	}
	return true
}

func (x *MinioDriver) ListObjects(bucketName string) []Object {
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

func (x *MinioDriver) StatObject(bucketName, objectName string) (Object, error) {
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

func (x *MinioDriver) DownloadURL(bucketName string, objectName string) (string, error) {
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
