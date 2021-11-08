package apilog

import "time"

type Entry struct {
	Request  Request       `json:"request"`
	Response Response      `json:"response"`
	Duration time.Duration `json:"duration"`
}

type Request struct {
	Method string            `json:"method,omitempty"`
	Size   int64             `json:"size,omitempty"`
	Url    Url               `json:"url"`
	Header map[string]string `json:"header,omitempty"`
}

type Url struct {
	Path string `json:"path,omitempty"`
	Host string `json:"host,omitempty"`
}

type Response struct {
	Code int `json:"code,omitempty"`
	Size int64 `json:"size,omitempty"`
}
