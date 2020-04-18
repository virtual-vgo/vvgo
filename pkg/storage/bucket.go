package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
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

func (x *Bucket) newSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	ctx, span := beeline.StartSpan(ctx, name)
	beeline.AddField(ctx, "bucket_name", x.Name)
	beeline.AddField(ctx, "bucket_region", x.Region)
	return ctx, span
}

func (x *Bucket) Make(ctx context.Context) bool {
	ctx, span := x.newSpan(ctx, "Bucket.Make")
	defer span.Send()
	exists, err := x.BucketExists(x.Name)
	if err != nil {
		logger.WithError(err).Error("minioClient.BucketExists() failed")
		span.AddField("error", err)
		return false
	}
	if exists == false {
		if err := x.MakeBucket(x.Name, x.Region); err != nil {
			logger.WithError(err).Error("minioClient.MakeBucket() failed")
			span.AddField("error", err)
			return false
		}
	}
	return true
}

func (x *Bucket) StatObject(ctx context.Context, objectName string) (Object, error) {
	ctx, span := x.newSpan(ctx, "Bucket.StatsObject")
	defer span.Send()
	opts := minio.StatObjectOptions{}
	objectInfo, err := x.Client.StatObject(x.Name, objectName, opts)
	if err != nil {
		span.AddField("error", err)
		return Object{}, err
	} else {
		return Object{
			ContentType: objectInfo.ContentType,
			Tags:        Tags(objectInfo.UserMetadata),
		}, nil
	}
}

func (x *Bucket) GetObject(ctx context.Context, name string, dest *Object) bool {
	ctx, span := x.newSpan(ctx, "Bucket.GetObject")
	defer span.Send()
	minioObject, err := x.Client.GetObject(x.Name, name, minio.GetObjectOptions{})
	if err != nil {
		logger.WithError(err).Error("minioClient.GetObject() failed")
		span.AddField("error", err)
		return false
	}
	info, err := minioObject.Stat()
	if err != nil {
		logger.WithError(err).Error("minioObject.Stat() failed")
		span.AddField("error", err)
		return false
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(minioObject); err != nil {
		logger.WithError(err).Error("minioObject.Read() failed")
		span.AddField("error", err)
		return false
	}
	*dest = Object{
		ContentType: info.ContentType,
		Tags:        map[string]string(info.UserMetadata),
		Buffer:      buffer,
	}
	return true
}

func (x *Bucket) PutObject(ctx context.Context, name string, object *Object) bool {
	ctx, span := x.newSpan(ctx, "Bucket.PutObject")
	defer span.Send()

	if ok := x.Make(ctx); !ok {
		return false
	}
	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.Client.PutObject(x.Name, name, &object.Buffer, -1, opts)
	if err != nil {
		logger.WithError(err).Error("minioClient.PutObject() failed")
		span.AddField("error", err)
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
func WithBackup(putObjectFunc func(ctx context.Context, name string, object *Object) bool) func(ctx context.Context, name string, object *Object) bool {
	return func(ctx context.Context, name string, object *Object) bool {
		backupName := fmt.Sprintf("%s-%s", name, time.Now().UTC().Format(time.RFC3339))
		backupBuffer := object.Buffer
		if ok := putObjectFunc(ctx, backupName, NewObject(object.ContentType, &backupBuffer)); !ok {
			return false
		}
		return putObjectFunc(ctx, name, object)
	}
}

func (x *Bucket) PutFile(ctx context.Context, file *File) bool {
	ctx, span := x.newSpan(ctx, "Bucket.PutFile")
	defer span.Send()
	return x.PutObject(ctx, file.ObjectKey(), NewObject(file.MediaType, bytes.NewBuffer(file.Bytes)))
}

func (x *Bucket) ListObjects(ctx context.Context) map[string]Object {
	ctx, span := x.newSpan(ctx, "Bucket.ListObjects")
	defer span.Send()
	done := make(chan struct{})
	defer close(done)

	objects := make(map[string]Object)
	for objectInfo := range x.Client.ListObjects(x.Name, "", false, done) {
		if objectInfo.Key == "" {
			continue
		}

		object, err := x.StatObject(ctx, objectInfo.Key)
		if err != nil {
			logger.WithError(err).Error("minio.StatObject() failed")
			continue
		}
		objects[objectInfo.Key] = object
	}
	return objects
}

func (x *Bucket) DownloadURL(ctx context.Context, name string) (string, error) {
	ctx, span := x.newSpan(ctx, "Bucket.DownloadURL")
	defer span.Send()
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
