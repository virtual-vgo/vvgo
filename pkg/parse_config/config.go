package parse_config

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"os"
	"reflect"
	"strconv"
)

var logger = log.New()

const DefaultConfigFile = "/etc/vvgo/vvgo.json"

type CtxKey string

const CtxKeyVVGOConfig CtxKey = "vvgo_config"

func (x CtxKey) Module(module string) CtxKey { return x + CtxKey("_"+module) }


func SetModuleConfig(ctx context.Context, module string, src interface{}) context.Context {
	return context.WithValue(ctx, CtxKeyVVGOConfig.Module(module), src)
}

func ReadConfigModule(ctx context.Context, module string, dest interface{}) {
	moduleData := ctx.Value(CtxKeyVVGOConfig.Module(module))
	if moduleData != nil {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(moduleData))
		return
	}

	configFile := os.Getenv("VVGO_CONFIGURATION_FILE")
	if configFile == "" {
		configFile = DefaultConfigFile
	}

	configJSON := make(map[string]json.RawMessage)
	file, err := os.Open(configFile)
	if err != nil {
		logger.SomeMethodFailure(ctx, "os.Open", err)
	} else {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&configJSON); err != nil {
			logger.JsonDecodeFailure(ctx, err)
		} else if moduleJSON, ok := configJSON[module]; ok {
			print(string(moduleJSON))
			if err := json.Unmarshal(moduleJSON, dest); err != nil {
				logger.JsonDecodeFailure(ctx, err)
			} else {
				return
			}
		}
	}

	logger.WithField("config_module", module).Errorf("config module `%s` not found", module)
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
