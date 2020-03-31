package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Package to query version information

var (
	version Version
	setOnce sync.Once
)

func Set(ver Version)       { setOnce.Do(func() { version = ver }) }
func String() string        { return version.String() }
func JSON() json.RawMessage { return version.JSON() }
func Header() http.Header   { return version.Header() }
func ReleaseTags() []string { return version.ReleaseTags() }

type Version struct {
	BuildHost string `json:"build_host"`
	BuildTime string `json:"build_time"`
	GitSha    string `json:"git_sha"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
}

func (x Version) String() string {
	return fmt.Sprintf("%s-%s", x.GitBranch, x.GitSha)
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

func (x Version) ReleaseTags() []string {
	return []string{
		x.GitBranch,
		fmt.Sprintf("%s-%s", x.GitBranch, x.GitSha),
	}
}
