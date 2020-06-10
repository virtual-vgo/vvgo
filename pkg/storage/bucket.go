package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

var logger = log.Logger()

var warehouse Warehouse

func Initialize(config Config) {
	minioClient, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, config.UseSSL)
	if err != nil {
		logger.WithError(err).Fatal("minio.New() failed")
	}
	warehouse = Warehouse{config: config, minioClient: minioClient}
}

// Warehouse builds new object storage buckets.
// Minio is the underlying driver.
type Warehouse struct {
	config      Config
	minioClient *minio.Client
}

type Config struct {
	Endpoint  string `default:"localhost:9000"`
	Region    string `default:"sfo2"`
	AccessKey string `default:"minioadmin"`
	SecretKey string `default:"minioadmin"`
	UseSSL    bool   `default:"false"`
}

// Buckets are an abstraction on top of the minio client for object storage
type Bucket struct {
	Name        string
	minioRegion string
	minioClient *minio.Client
}

// NewBucket returns a client for the bucket with name `name`, and creates the bucket if it does not exist.
func NewBucket(ctx context.Context, name string) (*Bucket, error) {
	return warehouse.NewBucket(ctx, name)
}

func (x *Warehouse) NewBucket(ctx context.Context, name string) (*Bucket, error) {
	bucket := Bucket{
		Name:        name,
		minioRegion: x.config.Region,
		minioClient: x.minioClient,
	}

	_, span := x.newSpan(ctx, "warehouse_new_bucket")
	defer span.Send()
	exists, err := x.minioClient.BucketExists(name)
	if err != nil {
		span.AddField("error", err)
		return nil, err
	}
	if exists == false {
		if err := x.minioClient.MakeBucket(name, x.config.Region); err != nil {
			span.AddField("error", err)
			return nil, err
		}
	}
	return &bucket, nil
}

func (x *Warehouse) newSpan(ctx context.Context, name string) (context.Context, tracing.Span) {
	ctx, span := tracing.StartSpan(ctx, name)
	tracing.AddField(ctx, "minio_endpoint", x.config.Endpoint)
	tracing.AddField(ctx, "minio_region", x.config.Region)
	return ctx, span
}

// File is an object abstraction for any media files, pdfs, mp3s, etc. that might be uploaded to the website.
// The object key is the md5sum of the file bytes.
// This should not be used for data files.
type File struct {
	// Mime media type
	MediaType string

	// File extension including the dot
	Ext string

	// File payload
	Bytes []byte

	// This is a cache of the objectKey
	objectKey string
}

var ErrInvalidMediaType = fmt.Errorf("invalid media type")
var ErrDetectedInvalidContent = fmt.Errorf("detected invalid content")
var ErrInvalidFileExtension = fmt.Errorf("invalid file extension")

// ValidateMediaType checks the media type for files in 3 ways:
// 1. whatever is set in x.MediaType,
// 2. using http.DetectContentType to sniff the fist 512 bytes,
// 3. and by the file extension.
// If any of these media types do not pre as a prefix, an error is returned.
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

// ObjectKey is the md5sum of the file bytes.
func (x File) ObjectKey() string {
	if x.objectKey == "" {
		x.objectKey = fmt.Sprintf("%x%s", md5.Sum(x.Bytes), x.Ext)
	}
	return x.objectKey
}

// StatFile queries object storage for the file media type and ext.
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

// PutFile uploads the file to object storage
func (x *Bucket) PutFile(ctx context.Context, file *File) error {
	ctx, span := x.newSpan(ctx, "bucket_put_file")
	defer span.Send()
	return x.PutObject(ctx, file.ObjectKey(), NewObject(file.MediaType, nil, file.Bytes))
}

// Objects are used as the container for storing persistent data.
// Data using this api should be able to serialize into bytes and have some associated media type.
// Objects can be stored in in-memory caches or remote object storage.
type Object struct {
	ContentType string
	Tags        map[string]string
	Bytes       []byte
}

func NewObject(mediaType string, tags map[string]string, payload []byte) *Object {
	return &Object{
		ContentType: mediaType,
		Tags:        tags,
		Bytes:       payload,
	}
}

type ObjectInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
	ContentType  string
}

func (x *Bucket) ListObjects(ctx context.Context, pre string) []ObjectInfo {
	done := make(chan struct{})
	defer close(done)
	_, span := x.newSpan(ctx, "bucket_list_objects")
	defer span.Send()
	var info []ObjectInfo
	for objectInfo := range x.minioClient.ListObjects(x.Name, pre, false, done) {
		info = append(info, ObjectInfo{
			Key:          objectInfo.Key,
			LastModified: objectInfo.LastModified,
			Size:         objectInfo.Size,
			ContentType:  objectInfo.ContentType,
		})
	}
	return info
}

// StatObject returns only the object content type and tags.
func (x *Bucket) StatObject(ctx context.Context, objectName string, dest *Object) error {
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
		Tags:        objectInfo.UserMetadata,
	}
	return nil
}

// GetObject returns the full object meta and contents.
func (x *Bucket) GetObject(ctx context.Context, name string, dest *Object) error {
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
	if _, err = buffer.ReadFrom(minioObject); err != nil {
		span.AddField("error", err)
		return err
	}

	*dest = Object{
		ContentType: info.ContentType,
		Tags:        info.UserMetadata,
		Bytes:       buffer.Bytes(),
	}
	return nil
}

// PutObject uploads the object with the given key.
func (x *Bucket) PutObject(ctx context.Context, name string, object *Object) error {
	ctx, span := x.newSpan(ctx, "bucket_put_object")
	defer span.Send()
	opts := minio.PutObjectOptions{
		ContentType:  object.ContentType,
		UserMetadata: object.Tags,
	}
	n, err := x.minioClient.PutObject(x.Name, name, bytes.NewBuffer(object.Bytes), -1, opts)
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

// DownloadURL queries object storage for a download url to the object key.
// If the object has a public download policy, then a direct link is returned.
// Otherwise, this method will query object storage for a presigned url.
func (x *Bucket) DownloadURL(ctx context.Context, name string) (string, error) {
	ctx, span := x.newSpan(ctx, "Bucket.DownloadURL")
	defer span.Send()
	downloadUrl, err := x.minioClient.PresignedGetObject(x.Name, name, ProtectedLinkExpiry, nil)
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
