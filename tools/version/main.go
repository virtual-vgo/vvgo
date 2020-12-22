package main

import (
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func main() {
	if file, err := os.OpenFile(version.VersionFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600); err != nil {
		panic(err)
	} else if err = json.NewEncoder(file).Encode(&version.Version{
		BuildHost: hostname(),
		BuildTime: time.Now().String(),
		GitSha:    gitSha(),
		GitBranch: gitBranch(),
		GoVersion: runtime.Version(),
	}); err != nil {
		panic(err)
	}
	log.Logger().WithField("path", version.VersionFile).Info("wrote version.json")
	os.Exit(0)
}

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(host)
}

func gitSha() string {
	if os.Getenv("GIT_COMMIT") != "" {
		return os.Getenv("GIT_COMMIT")
	}
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "command `git rev-parse HEAD` failed!: %v\n", err)
	}
	return strings.TrimSpace(string(output))
}

func gitBranch() string {
	if os.Getenv("GIT_BRANCH") != "" {
		return os.Getenv("GIT_BRANCH")
	}
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "command `git rev-parse --abbrev-ref HEAD` failed!%v\n", err)
	}
	return strings.TrimSpace(string(output))
}
