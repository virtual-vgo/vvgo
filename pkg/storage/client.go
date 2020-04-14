package storage

import (
	"github.com/go-redis/redis/v7"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"sync"
	"time"
)

const RedisLockDeadline = 5 * 60 * time.Second
const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

var logger = log.Logger()

type Client struct {
	config      Config
	minioClient *minio.Client
	redisClient *redis.Client

	lockers map[string]*Locker
	lock    sync.Mutex
}

type Config struct {
	Minio MinioConfig `envconfig:"minio"`
	Redis RedisConfig `envconfig:"redis"`
}

type MinioConfig struct {
	Endpoint  string `default:"localhost:9000"`
	Region    string `default:"sfo2"`
	AccessKey string `default:"minioadmin"`
	SecretKey string `default:"minioadmin"`
	UseSSL    bool   `default:"false"`
}

type RedisConfig struct {
	Address string `default:"localhost:6379"`
}

func NewClient(config Config) *Client {
	minioClient, err := minio.New(config.Minio.Endpoint, config.Minio.AccessKey, config.Minio.SecretKey, config.Minio.UseSSL)
	if err != nil {
		logger.WithError(err).Error("minio.New() failed")
		return nil
	}
	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.Redis.Address,
	})

	return &Client{
		config:      config,
		minioClient: minioClient,
		redisClient: redisClient,
		lockers:     make(map[string]*Locker),
	}
}
