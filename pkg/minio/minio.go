package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"reflect"
	"strconv"
)

var logger = log.Logger()

type Config struct {
	Endpoint  string `redis:"endpoint" default:"localhost:9000"`
	Region    string `redis:"region" default:"sfo2"`
	AccessKey string `redis:"access_key" default:"minioadmin"`
	SecretKey string `redis:"secret_key" default:"minioadmin"`
	UseSSL    bool   `redis:"use_ssl" default:"false"`
}

func (x *Config) readRedis(ctx context.Context) error {
	redisVals := make(map[string]string)
	err := redis.Do(ctx, redis.Cmd(&redisVals, "HGETALL", "config:minio"))
	if err != nil {
		return err
	}

	reflectType := reflect.TypeOf(x).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		redisKey := field.Tag.Get("redis")
		if redisKey == "" {
			continue
		}
		if _, ok := redisVals[redisKey]; !ok {
			continue
		}
		setField(x, field.Type.Kind(), i, redisVals[redisKey])
	}
	return nil
}

func (x *Config) readDefaults() {
	reflectType := reflect.TypeOf(x).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		defaultString := field.Tag.Get("default")
		if defaultString == "" {
			continue
		}
		setField(x, field.Type.Kind(), i, defaultString)
	}
}

func NewClient(ctx context.Context) (*minio.Client, error) {
	var config Config
	config.readDefaults()
	if err := config.readRedis(ctx); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
	}

	minioClient, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, config.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("minio.New() failed: %w", err)
	}
	return minioClient, nil
}

func setField(dest interface{}, kind reflect.Kind, i int, valString string) {
	switch kind {
	case reflect.String:
		reflect.ValueOf(dest).Elem().Field(i).SetString(valString)
	case reflect.Bool:
		val, _ := strconv.ParseBool(valString)
		reflect.ValueOf(dest).Elem().Field(i).SetBool(val)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		val, _ := strconv.ParseInt(valString, 10, 64)
		reflect.ValueOf(dest).Elem().Field(i).SetInt(val)
	}
}
