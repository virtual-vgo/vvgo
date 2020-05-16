package version

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Package to query version information

var version Version

func Set(ver Version)       { version = ver }
func String() string        { return version.String() }
func JSON() json.RawMessage { return version.JSON() }
func Header() http.Header   { return version.Header() }

type Version struct {
	BuildHost string `json:"build_host"`
	BuildTime string `json:"build_time"`
	GitSha    string `json:"git_sha"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
}

func (x Version) String() string {
	return fmt.Sprintf("%s", x.GitSha)
}

func (x Version) JSON() json.RawMessage {
	jsonBytes, _ := json.Marshal(x)
	return jsonBytes
}

func (x Version) Header() http.Header {
	header := make(http.Header)
	header.Set("Build-Host", x.BuildHost)
	header.Set("Build-Time", x.BuildTime)
	header.Set("Git-Sha", x.GitSha)
	header.Set("Git-Branch", x.GitBranch)
	header.Set("Go-Version", x.GoVersion)
	return header
}
