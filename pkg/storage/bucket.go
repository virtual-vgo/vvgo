package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

var logger = log.Logger()
var client *Warehouse

func init() {
	var config Config
	err := envconfig.Process("storage", &config)
	if err != nil {
		logger.Fatal(err)
	}
	client, err = NewWarehouse(config)
	if err != nil {
		logger.Fatal(err)
	}
}

type Warehouse struct {
	Config
	minioClient *minio.Client
}

type Config struct {
	Minio MinioConfig `envconfig:"minio"`
}

type MinioConfig struct {
	Endpoint  string `default:"localhost:9000"`
	Region    string `default:"sfo2"`
	AccessKey string `default:"minioadmin"`
	SecretKey string `default:"minioadmin"`
	UseSSL    bool   `default:"false"`
}

func NewWarehouse(config Config) (*Warehouse, error) {
	client := Warehouse{Config: config}
	if config.Minio.Endpoint != "" {
		var err error
		client.minioClient, err = minio.New(config.Minio.Endpoint, config.Minio.AccessKey, config.Minio.SecretKey, config.Minio.UseSSL)
		if err != nil {
			return nil, fmt.Errorf("minio.New() failed: %v", err)
		}
	}
	return &client, nil
}

func NewBucket(ctx context.Context, name string) (*Bucket, error) {
	return client.NewBucket(ctx, name)
}

type Bucket struct {
	Name        string
	minioRegion string
	minioClient *minio.Client
}

func (x *Warehouse) NewBucket(ctx context.Context, name string) (*Bucket, error) {
	bucket := Bucket{
		Name:        name,
		minioRegion: x.Minio.Region,
		minioClient: x.minioClient,
	}

	if x.minioClient != nil {
		bucket.minioClient = x.minioClient
		_, span := x.newSpan(ctx, "warehouse_new_bucket")
		defer span.Send()
		exists, err := x.minioClient.BucketExists(name)
		if err != nil {
			span.AddField("error", err)
			return nil, err
		}
		if exists == false {
			if err := x.minioClient.MakeBucket(name, x.Minio.Region); err != nil {
				span.AddField("error", err)
				return nil, err
			}
		}
	}
	return &bucket, nil
}

func (x *Warehouse) newSpan(ctx context.Context, name string) (context.Context, tracing.Span) {
	ctx, span := tracing.StartSpan(ctx, name)
	tracing.AddField(ctx, "minio_endpoint", x.Minio.Endpoint)
	tracing.AddField(ctx, "minio_region", x.Minio.Region)
	return ctx, span
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

func (x *Bucket) StatFile(ctx context.Context, objectKey string, dest *File) error {
	var obj Object
	if err := x.StatObject(ctx, objectKey, &obj); err != nil {
		return err
	}
	*dest = File{
		Ext:       filepath.Ext(objectKey),
		MediaType: obj.ContentType,
		objectKey: objectKey,
	}
	return nil
}

func (x *Bucket) PutFile(ctx context.Context, file *File) error {
	ctx, span := x.newSpan(ctx, "bucket_put_file")
	defer span.Send()
	return x.PutObject(ctx, file.ObjectKey(), NewObject(file.MediaType, bytes.NewBuffer(file.Bytes)))
}

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

func (x *Bucket) StatObject(ctx context.Context, objectName string, dest *Object) error {
	if x.minioClient == nil {
		return nil
	}

	_, span := x.newSpan(ctx, "bucket_stat_object")
	defer span.Send()
	opts := minio.StatObjectOptions{}
	objectInfo, err := x.minioClient.StatObject(x.Name, objectName, opts)
	if err != nil {
		span.AddField("error", err)
		return err
	}
	*dest = Object{
		ContentType: objectInfo.ContentType,
		Tags:        Tags(objectInfo.UserMetadata),
	}
	return nil
}

func (x *Bucket) GetObject(ctx context.Context, name string, dest *Object) error {
	if x.minioClient == nil {
		return nil
	}

	_, span := x.newSpan(ctx, "bucket_get_object")
	defer span.Send()
	minioObject, err := x.minioClient.GetObject(x.Name, name, minio.GetObjectOptions{})
	if err != nil {
		span.AddField("error", err)
		return err
	}
	info, err := minioObject.Stat()
	if err != nil {
		span.AddField("error", err)
		return err
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(minioObject); err != nil {
		span.AddField("error", err)
		return err
	}
	*dest = Object{
		ContentType: info.ContentType,
		Tags:        map[string]string(info.UserMetadata),
		Buffer:      buffer,
	}
	return err
}

func (x *Bucket) PutObject(ctx context.Context, name string, object *Object) error {
	if x.minioClient == nil {
		return nil
	}

	ctx, span := x.newSpan(ctx, "bucket_put_object")
	defer span.Send()
	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.minioClient.PutObject(x.Name, name, &object.Buffer, -1, opts)
	if err != nil {
		span.AddField("error", err)
		return err
	}
	logger.WithFields(logrus.Fields{
		"bucket_name": x.Name,
		"object_name": name,
		"object_size": n,
	}).Info("uploaded object")
	return nil
}

// Stores the object and a copy with a timestamp appended to the file name.
func WithBackup(putObjectFunc func(ctx context.Context, name string, object *Object) error) func(ctx context.Context, name string, object *Object) error {
	return func(ctx context.Context, name string, object *Object) error {
		backupName := fmt.Sprintf("%s-%s", name, time.Now().UTC().Format(time.RFC3339))
		backupBuffer := object.Buffer
		if err := putObjectFunc(ctx, backupName, NewObject(object.ContentType, &backupBuffer)); err != nil {
			return err
		}
		return putObjectFunc(ctx, name, object)
	}
}

func (x *Bucket) DownloadURL(ctx context.Context, name string) (string, error) {
	if x.minioClient == nil {
		return "#", nil
	}

	ctx, span := x.newSpan(ctx, "Bucket.DownloadURL")
	defer span.Send()
	policy, err := x.minioClient.GetBucketPolicy(x.Name)
	if err != nil {
		return "", err
	}

	var downloadUrl *url.URL
	switch policy {
	case "download":
		downloadUrl, err = url.Parse(fmt.Sprintf("%s/%s/%s", x.minioClient.EndpointURL(), x.Name, name))
	default:
		downloadUrl, err = x.minioClient.PresignedGetObject(x.Name, name, ProtectedLinkExpiry, nil)
	}
	if err != nil {
		return "", err
	} else {
		return downloadUrl.String(), nil
	}
}

func (x *Bucket) newSpan(ctx context.Context, name string) (context.Context, tracing.Span) {
	ctx, span := tracing.StartSpan(ctx, name)
	tracing.AddField(ctx, "bucket_name", x.Name)
	tracing.AddField(ctx, "bucket_region", x.minioRegion)
	return ctx, span
}
