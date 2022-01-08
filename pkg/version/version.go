package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Package to query version information

const FileName = "version.json"

var version Version

func init() {
	file, err := os.Open(FileName)
	if err != nil {
		return
	}
	defer file.Close()
	_ = json.NewDecoder(file).Decode(&version)
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

func BuildTime() time.Time {
	return version.BuildTime
}

type Version struct {
	BuildTime time.Time `json:"build_time"`
	GitSha    string    `json:"git_sha"`
	GoVersion string    `json:"go_version"`
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
	header.Set("Build-Time", x.BuildTime.String())
	header.Set("Git-Sha", x.GitSha)
	header.Set("Go-Version", x.GoVersion)
	return header
}
