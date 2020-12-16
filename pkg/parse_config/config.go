package parse_config

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"reflect"
	"strconv"
)

func ReadFromRedisHash(ctx context.Context, dest interface{}, redisKey string) error {
	redisVals := make(map[string]string)
	err := redis.Do(ctx, redis.Cmd(&redisVals, "HGETALL", redisKey))
	if err != nil {
		return err
	}

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

func DistroBucket(ctx context.Context) string {
	var distroBucket string
	redis.Do(ctx, redis.Cmd(&distroBucket, "GET", "config:distro_bucket"))
	return distroBucket
}
