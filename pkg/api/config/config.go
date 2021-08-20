package config

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
)

var logger = log.New()
var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		helpers.MethodNotAllowed(w)
	}
	if err := json.NewEncoder(w).Encode(&parse_config.Config); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
})
