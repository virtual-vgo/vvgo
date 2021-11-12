package http_util

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/http/httputil"
)

func NoFollow(client *http.Client) *http.Client {
	if client == nil {
		client = new(http.Client)
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(req)
}

func DoRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var respErr error

	defer func() {
		if config.Env.DebugHTTP {
			sendBytes, _ := httputil.DumpRequestOut(req, true)
			recvBytes, _ := httputil.DumpResponse(resp, true)
			fmt.Println("--- BEGIN DEBUG HTTP CLIENT REQUEST ---")
			fmt.Println("Sent request:")
			fmt.Println(string(sendBytes))
			fmt.Println("Received response:")
			fmt.Println(string(recvBytes))
			fmt.Println("--- END DEBUG HTTP CLIENT REQUEST ---")
		}
	}()

	span, spanOk := tracing.NewSpanFromContext(req.Context(), "outgoing http request")
	if !spanOk {
		logger.Warn("http client: invalid trace context")
	}

	resp, respErr = http.DefaultClient.Do(req)
	requestMetrics := tracing.NewHttpRequestMetrics(req)
	responseMetrics := tracing.NewHttpResponseMetrics(resp.StatusCode, resp.ContentLength)
	if spanOk {
		tracing.WriteSpan(
			span.Finish().
				WithHttpRequestMetrics(requestMetrics).
				WithHttpResponseMetrics(responseMetrics).
				WithError(respErr),
		)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		logger.
			WithFields(requestMetrics.Fields()).
			WithFields(responseMetrics.Fields()).
			Info("http client: request completed")
	default:
		logger.
			WithFields(requestMetrics.Fields()).
			WithFields(responseMetrics.Fields()).
			Warn("http client: request completed with error status")
	}
	return resp, respErr
}



