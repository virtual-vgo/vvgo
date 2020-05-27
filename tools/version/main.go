package main

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"
)

var infoTemplate = template.Must(template.New("info.go").Parse(`
package main

import "github.com/virtual-vgo/vvgo/pkg/version"

func init() {
 	version.Set(version.Version{
		BuildHost: "{{ .BuildHost }}",
		BuildTime: "{{ .BuildTime }}",
		GitSha:    "{{ .GitSha }}",
		GitBranch: "{{ .GitBranch }}",
		GoVersion: "{{ .GoVersion }}",
	})
}`))

func main() {
	ver := version.Version{
		BuildHost: hostname(),
		BuildTime: time.Now().String(),
		GitSha:    gitSha(),
		GitBranch: gitBranch(),
		GoVersion: runtime.Version(),
	}
	os.Remove("info.go")
	output, err := os.OpenFile("info.go", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	writeVersion(output, ver)
}

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(host)
}

func gitSha() string {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "command `git rev-parse HEAD` failed!")
		panic(err)
	}
	return strings.TrimSpace(string(output))
}

func gitBranch() string {
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "command `git rev-parse --abbrev-ref HEAD` failed!")
		panic(err)
	}
	return strings.TrimSpace(string(output))
}

// writes the info to the io writer
// output is passed through gofmt
func writeVersion(output io.Writer, ver version.Version) {
	if err := func() error {
		// use gofmt to actually write the file
		gofmt := exec.Command("gofmt")
		gofmt.Stdout = output
		gofmt.Stderr = os.Stderr

		// make an input pipe for gofmt
		gofmtIn, err := gofmt.StdinPipe()
		if err != nil {
			return fmt.Errorf("cmd.StdinPipe() failed: %v", err)
		}

		// start gofmt
		err = gofmt.Start()
		if err != nil {
			return fmt.Errorf("cmd.Start() failed: %v", err)
		}

		// run the template
		err = infoTemplate.Execute(gofmtIn, ver)
		if err != nil {
			return fmt.Errorf("template.Execute() failed: %v", err)
		}

		// close the input pipe
		if err = gofmtIn.Close(); err != nil {
			return fmt.Errorf("gofmtIn.Close() failed: %v", err)
		}

		// check that gofmt exits success
		if err = gofmt.Wait(); err != nil {
			return fmt.Errorf("cmd.Wait() failed: %v", err)
		}
		return nil
	}(); err != nil {
		panic(err)
	}
}
