package storage

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"net/url"
)

type Bucket struct {
	Name   string
	Region string
	*minio.Client
}

type Object struct {
	ContentType string
	Tags        Tags
	Buffer      bytes.Buffer
}

type Tags map[string]string

func (x *Client) NewBucket(name string) *Bucket {
	return &Bucket{
		Name:   name,
		Region: x.config.MinioConfig.Region,
		Client: x.minioClient,
	}
}

func (x *Bucket) Make() bool {
	exists, err := x.BucketExists(x.Name)
	if err != nil {
		logger.WithError(err).Error("minioClient.BucketExists() failed")
		return false
	}
	if exists == false {
		if err := x.MakeBucket(x.Name, x.Region); err != nil {
			logger.WithError(err).Error("minioClient.MakeBucket() failed")
			return false
		}
	}
	return true
}

func (x *Bucket) StatObject(objectName string) (Object, error) {
	opts := minio.StatObjectOptions{}
	objectInfo, err := x.Client.StatObject(x.Name, objectName, opts)
	if err != nil {
		return Object{}, err
	} else {
		return Object{
			ContentType: objectInfo.ContentType,
			Tags:        Tags(objectInfo.UserMetadata),
		}, nil
	}
}

func (x *Bucket) GetObject(name string, dest *Object) bool {
	minioObject, err := x.Client.GetObject(x.Name, name, minio.GetObjectOptions{})
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
		Tags:        map[string]string(info.UserMetadata),
		Buffer:      buffer,
	}
	return true
}

func (x *Bucket) PutObject(name string, object *Object) bool {
	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.Client.PutObject(x.Name, name, &object.Buffer, -1, opts)
	if err != nil {
		logger.WithError(err).Error("minioClient.PutObject() failed")
		return false
	}
	logger.WithFields(logrus.Fields{
		"object_name": name,
		"object_size": n,
	}).Info("uploaded object")
	return true
}

func (x *Bucket) ListObjects() map[string]Object {
	done := make(chan struct{})
	defer close(done)

	objects := make(map[string]Object)
	for objectInfo := range x.Client.ListObjects(x.Name, "", false, done) {
		if objectInfo.Key == "" {
			continue
		}

		object, err := x.StatObject(objectInfo.Key)
		if err != nil {
			logger.WithError(err).Error("minio.StatObject() failed")
			continue
		}
		objects[objectInfo.Key] = object
	}
	return objects
}

func (x *Bucket) DownloadURL(name string) (string, error) {
	policy, err := x.Client.GetBucketPolicy(x.Name)
	if err != nil {
		return "", err
	}

	var downloadUrl *url.URL
	switch policy {
	case "download":
		downloadUrl, err = url.Parse(fmt.Sprintf("%s/%s/%s", x.Client.EndpointURL(), x.Name, name))
	default:
		downloadUrl, err = x.Client.PresignedGetObject(x.Name, name, ProtectedLinkExpiry, nil)
	}
	if err != nil {
		return "", err
	} else {
		return downloadUrl.String(), nil
	}
}
