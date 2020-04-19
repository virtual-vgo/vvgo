package storage

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Object struct {
	ContentType string
	Tags        Tags
	Buffer      bytes.Buffer
}

type Tags map[string]string

func NewObject(mediaType string, buffer *bytes.Buffer) *Object {
	return &Object{
		ContentType: mediaType,
		Buffer:      *buffer,
	}
}

func NewJSONObject(buffer *bytes.Buffer) *Object {
	return NewObject("application/json", buffer)
}

type File struct {
	MediaType string
	Ext       string
	Bytes     []byte
	objectKey string
}

var ErrInvalidMediaType = fmt.Errorf("invalid media type")
var ErrDetectedInvalidContent = fmt.Errorf("detected invalid content")
var ErrInvalidFileExtension = fmt.Errorf("invalid file extension")

func (x File) ValidateMediaType(pre string) error {
	switch {
	case !strings.HasPrefix(x.MediaType, pre):
		return ErrInvalidMediaType
	case !strings.HasPrefix(http.DetectContentType(x.Bytes), pre):
		return ErrDetectedInvalidContent
	case !strings.HasPrefix(mime.TypeByExtension(x.Ext), pre):
		return ErrInvalidFileExtension
	default:
		return nil
	}
}

func (x File) ObjectKey() string {
	if x.objectKey == "" {
		x.objectKey = fmt.Sprintf("%x%s", md5.Sum(x.Bytes), x.Ext)
	}
	return x.objectKey
}

type Bucket struct {
	Name   string
	Region string
	*minio.Client
}

func (x *Client) NewBucket(name string) *Bucket {
	return &Bucket{
		Name:   name,
		Region: x.config.Minio.Region,
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

func (x *Bucket) StatFile(file *File) (Object, error) {
	return x.StatObject(file.ObjectKey())
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
	if ok := x.Make(); !ok {
		return false
	}
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
		"bucket_name": x.Name,
		"object_name": name,
		"object_size": n,
	}).Info("uploaded object")
	return true
}

// Stores the object and a copy with a timestamp appended to the file name.
func WithBackup(putObjectFunc func(name string, object *Object) bool) func(name string, object *Object) bool {
	return func(name string, object *Object) bool {
		backupName := fmt.Sprintf("%s-%s", name, time.Now().UTC().Format(time.RFC3339))
		backupBuffer := object.Buffer
		if ok := putObjectFunc(backupName, NewObject(object.ContentType, &backupBuffer)); !ok {
			return false
		}
		return putObjectFunc(name, object)
	}
}

func (x *Bucket) PutFile(file *File) bool {
	return x.PutObject(file.ObjectKey(), NewObject(file.MediaType, bytes.NewBuffer(file.Bytes)))
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
