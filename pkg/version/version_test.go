package version

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sort"
	"strings"
	"testing"
)

func TestHeader(t *testing.T) {
	version = Version{
		BuildHost: "tuba-international.xyz",
		BuildTime: "today",
		GitSha:    "yeet",
		GitBranch: "best-branch",
		GoVersion: "1.14.1",
	}
	wantHeader := http.Header{
		"Build-Host": []string{"tuba-international.xyz"},
		"Build-Time": []string{"today"},
		"Git-Sha":    []string{"yeet"},
		"Git-Branch": []string{"best-branch"},
		"Go-Version": []string{"1.14.1"},
	}
	gotHeader := Header()
	assert.Equal(t, wantHeader, gotHeader)
}

func headerToString(header http.Header) string {
	var buf bytes.Buffer
	header.Write(&buf)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func TestJSON(t *testing.T) {
	version = Version{
		BuildHost: "tuba-international.xyz",
		BuildTime: "today",
		GitSha:    "yeet",
		GitBranch: "best-branch",
		GoVersion: "1.14.1",
	}
	wantJSON := `{"build_host":"tuba-international.xyz","build_time":"today","git_sha":"yeet","git_branch":"best-branch","go_version":"1.14.1"}`
	gotJSON := string(JSON())
	if expected, got := wantJSON, gotJSON; expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}

func TestSet(t *testing.T) {
	version = Version{} // reset
	Set(Version{
		BuildHost: "tuba-international.xyz",
		BuildTime: "today",
		GitSha:    "yeet",
		GitBranch: "best-branch",
		GoVersion: "1.14.1",
	})

	wantVersion := Version{
		BuildHost: "tuba-international.xyz",
		BuildTime: "today",
		GitSha:    "yeet",
		GitBranch: "best-branch",
		GoVersion: "1.14.1",
	}
	gotVersion := version

	if expected, got := fmt.Sprintf("%#v", wantVersion), fmt.Sprintf("%#v", gotVersion); expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}

func TestString(t *testing.T) {
	version = Version{
		BuildHost: "tuba-international.xyz",
		BuildTime: "today",
		GitSha:    "yeet",
		GitBranch: "best-branch",
		GoVersion: "1.14.1",
	}
	wantString := "yeet"
	gotString := String()

	if expected, got := wantString, gotString; expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}
