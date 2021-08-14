package parse_config

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
)

// ListenAddress Listen address for the http server.
var ListenAddress = "localhost:8080"

// ServerURL Url to reach the http server.
var ServerURL = "http://localhost:8080"

// FileName Path to the configuration file.
var FileName = "/etc/vvgo/vvgo.json"

// Endpoint RPC endpoint for remote configuration.
var Endpoint = "https://vvgo.org/api/v1/config"

// Session The session key returned by https://vvgo.org/api/v1/session?with_roles=read_config.
var Session string

var logger = log.New()

type CtxKey string

const CtxKeyVVGOConfig CtxKey = "vvgo_config"

func (x CtxKey) Module(module string) CtxKey { return x + CtxKey("_"+module) }

func SetModule(ctx context.Context, module string, src interface{}) context.Context {
	return context.WithValue(ctx, CtxKeyVVGOConfig.Module(module), src)
}

func ReadModule(ctx context.Context, module string, dest interface{}) {
	moduleData := ctx.Value(CtxKeyVVGOConfig.Module(module))
	switch {
	//case Session != "": TODO: Finish implementing this or delete it.
	//	ReadModuleFromEndpoint(ctx, Endpoint, Session, module, dest)

	case moduleData != nil:
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(moduleData))
	default:
		ReadModuleFromFile(ctx, FileName, module, dest)
	}
}

func ReadModuleFromFile(ctx context.Context, fileName string, module string, dest interface{}) {
	logger.Infof("reading config from %s", fileName)

	file, err := os.Open(FileName)
	if err != nil {
		logger.MethodFailure(ctx, "os.Open", err)
		return
	}
	defer file.Close()

	configJSON := make(map[string]json.RawMessage)
	if err := json.NewDecoder(file).Decode(&configJSON); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return
	}

	moduleJSON, ok := configJSON[module]
	if !ok {
		logger.WithField("config_module", module).Errorf("config module `%s` not found", module)
		return
	}

	if err := json.Unmarshal(moduleJSON, dest); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return
	}
}

func ReadModuleFromEndpoint(ctx context.Context, endpoint string, session string, module string, dest interface{}) {
	logger.Infof("fetching remote config from %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		logger.MethodFailure(ctx, "http.NewRequest", err)
		return
	}

	form := make(url.Values)
	form.Add("module", module)
	req.Header.Add("Authorization", "Bearer "+session)
	req.Form = form
	resp, err := http_wrappers.DoRequest(req)
	if err != nil {
		logger.MethodFailure(ctx, "http.Do", err)
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return
	}
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
