package parse_config

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

var Namespace = "config"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func UseTestNamespace() {
	Namespace = "test:config:" + strconv.Itoa(seededRand.Int())
}

func WriteRedisHashValue(ctx context.Context, redisKey, hashKey, val string) error {
	return redis.Do(ctx, redis.Cmd(nil, "HSET", Namespace+":"+redisKey, hashKey, val))
}

func WriteToRedisHash(ctx context.Context, redisKey string, src interface{}) error {
	var hashVals map[string]string

	switch val := src.(type) {
	case *map[string]string:
		hashVals = *val
	case map[string]string:
		hashVals = val
	default:
		hashVals = make(map[string]string)
		reflectType := reflect.TypeOf(src).Elem()
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			hashKey := field.Tag.Get("redis")
			if hashKey == "" {
				continue
			}
			hashVals[hashKey] = getField(src, field.Type.Kind(), i)
		}
	}
	args := make([]string, 1, 1+2*len(hashVals))
	args[0] = Namespace + ":" + redisKey
	for k, v := range hashVals {
		args = append(args, k, v)
	}
	return redis.Do(ctx, redis.Cmd(nil, "HMSET", args...))
}

func getField(src interface{}, kind reflect.Kind, i int) string {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(src).Elem().Field(i).String()
	case reflect.Bool:
		val := reflect.ValueOf(src).Elem().Field(i).Bool()
		return strconv.FormatBool(val)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		val := reflect.ValueOf(src).Elem().Field(i).Int()
		return strconv.FormatInt(val, 10)
	default:
		return ""
	}
}

func ReadRedisHashValue(ctx context.Context, redisKey string, hashKey string, dest *string) error {
	return redis.Do(ctx, redis.Cmd(dest, "HGET", Namespace+":"+redisKey, hashKey))
}

func ReadFromRedisHash(ctx context.Context, redisKey string, dest interface{}) error {
	redisVals := make(map[string]string)
	err := redis.Do(ctx, redis.Cmd(&redisVals, "HGETALL", Namespace+":"+redisKey))
	if err != nil {
		return err
	}

	switch dest.(type) {
	case *map[string]string:
		*dest.(*map[string]string) = redisVals
	default:
		reflectType := reflect.TypeOf(dest).Elem()
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			redisKey := field.Tag.Get("redis")
			if redisKey == "" {
				continue
			}
			if _, ok := redisVals[redisKey]; !ok {
				continue
			}
			setField(dest, field.Type.Kind(), i, redisVals[redisKey])
		}
	}
	return nil
}

func SetDefaults(dest interface{}) {
	reflectType := reflect.TypeOf(dest).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		defaultString := field.Tag.Get("default")
		if defaultString == "" {
			continue
		}
		if reflect.ValueOf(dest).Elem().Field(i).IsZero() {
			setField(dest, field.Type.Kind(), i, defaultString)
		}
	}
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
