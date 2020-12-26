package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"math/rand"
	"strconv"
	"time"
)

var logger = log.Logger()

type Config struct {
	Endpoint  string `redis:"endpoint" default:"localhost:9000"`
	Region    string `redis:"region" default:"sfo2"`
	AccessKey string `redis:"access_key" default:"minioadmin"`
	SecretKey string `redis:"secret_key" default:"minioadmin"`
	UseSSL    bool   `redis:"use_ssl" default:"false"`
}

func newConfig(ctx context.Context) Config {
	var dest Config
	parse_config.SetDefaults(&dest)
	if err := parse_config.ReadFromRedisHash(ctx, "minio", &dest); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
	}
	return dest
}

type Client struct{ minio.Client }

func NewClient(ctx context.Context) (*Client, error) {
	config := newConfig(ctx)
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
