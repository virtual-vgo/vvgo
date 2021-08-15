package minio

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"math/rand"
	"strconv"
	"time"
)

type Config struct {
	Endpoint  string `json:"endpoint" default:"localhost:9000"`
	AccessKey string `json:"access_key" default:"minioadmin"`
	SecretKey string `json:"secret_key" default:"minioadmin"`
	UseSSL    bool   `json:"use_ssl" default:"false"`
}

const ConfigModule = "minio"

type Client struct{ minio.Client }

func NewClient(ctx context.Context) (*Client, error) {
	var config Config
	if envconfig.Process("MINIO", &config) != nil { // Honor env vars over others
		parse_config.ReadModule(ctx, ConfigModule, &config)
		parse_config.SetDefaults(&config)
	}
	minioClient, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, config.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("minio.New() failed: %w", err)
	}
	return &Client{*minioClient}, nil
}

// Testing only!

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (x *Client) NewRandomBucket() (string, error) {
	bucketName := "bucket-" + strconv.Itoa(seededRand.Int())
	if err := x.MakeBucket(bucketName, "local"); err != nil {
		return "", err
	}
	return bucketName, nil
}
