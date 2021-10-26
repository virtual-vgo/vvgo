package version

import (
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"os"
)

// Package to query version information

const VersionFile = "version.json"

var version Version

func init() {
	if file, err := os.Open(VersionFile); err != nil {
		logger.WithError(err).WithField("path", VersionFile).Info("failed to open version file")
	} else if err = json.NewDecoder(file).Decode(&version); err != nil {
		logger.WithError(err).WithField("path", VersionFile).Info("failed unmarshal version file")
	}
}

func Get() Version          { return version }
func Set(ver Version)       { version = ver }
func String() string        { return version.String() }
func JSON() json.RawMessage { return version.JSON() }
func Header() http.Header   { return version.Header() }
func SetVersionHeaders(w http.ResponseWriter) {
	versionHeader := version.Header()
	for k := range versionHeader {
		w.Header().Set(k, versionHeader.Get(k))
	}
}

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
