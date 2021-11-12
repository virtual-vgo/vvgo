package minio

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"math/rand"
	"strconv"
	"time"
)

type Client struct{ minio.Client }

func NewClient() (*Client, error) {
	config := config.Env.Minio
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
