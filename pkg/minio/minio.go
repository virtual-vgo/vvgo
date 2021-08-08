package minio

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"math/rand"
	"strconv"
	"time"
)

var logger = log.New()

type Config struct {
	Endpoint  string `default:"localhost:9000"`
	Region    string `default:"sfo2"`
	AccessKey string `default:"minioadmin"`
	SecretKey string `default:"minioadmin"`
	UseSSL    bool   `default:"false"`
}

type Client struct{ minio.Client }

func NewClient(ctx context.Context) (*Client, error) {
	var config Config
	envconfig.MustProcess("MINIO", &config)
	minioClient, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, config.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("minio.New() failed: %w", err)
	}
	return &Client{*minioClient}, nil
}

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (x *Client) NewRandomBucket() (string, error) {
	bucketName := "bucket-" + strconv.Itoa(seededRand.Int())
	if err := x.MakeBucket(bucketName, "local"); err != nil {
		return "", err
	}
	return bucketName, nil
}
