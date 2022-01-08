package cloudflare

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"net/http"
	"strings"
)

func PurgeCache() {
	if config.Config.Cloudflare.ZoneId != "" && config.Config.Cloudflare.ApiKey != "" {
		req, err := http.NewRequest(http.MethodPost,
			"https://api.cloudflare.com/client/v4/zones/"+config.Config.Cloudflare.ZoneId+"/purge_cache",
			strings.NewReader(`{"purge_everything":true}`))
		if err != nil {
			log.WithError(err).Error("http.NewRequest() failed")
			log.Error("cloudflare cache purge failed")
			return
		}
		req.Header.Add("Authorization", "Bearer "+config.Config.Cloudflare.ApiKey)
		resp, err := http_wrappers.DoRequest(req)
		if err != nil {
			log.WithError(err).Error("http.Do() failed")
			log.Error("cloudflare cache purge failed")
			return
		}
		if resp.StatusCode != 200 {
			var body bytes.Buffer
			_, _ = body.ReadFrom(resp.Body)
			log.WithField("response_body", body.String()).Error("non-200 response from cloudflare")
			log.Error("cloudflare cache purge failed")
			return
		}
		log.Info("purged cloudflare cache")
	}
}
