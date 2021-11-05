package main

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
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
		BuildTime: time.Now().String(),
		GitSha:    gitSha(),
		GoVersion: runtime.Version(),
	}); err != nil {
		panic(err)
	}
	logger.New().WithField("path", version.VersionFile).Info("wrote version.json")
	os.Exit(0)
}

func gitSha() string {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		logger.New().Fatalf("command `git rev-parse HEAD` failed!: %v\n", err)
	}
	return strings.TrimSpace(string(output))
}
