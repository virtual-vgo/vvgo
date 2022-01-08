package version

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {
	version = Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	}
	wantHeader := http.Header{
		"Build-Time": []string{"2022-01-08 02:50:01 +0000 UTC"},
		"Git-Sha":    []string{"yeet"},
		"Go-Version": []string{"1.14.1"},
	}
	gotHeader := Header()
	assert.Equal(t, wantHeader, gotHeader)
}

func TestSetVersionHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	version = Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	}

	wantHeader := http.Header{
		"Build-Time": []string{"2022-01-08 02:50:01 +0000 UTC"},
		"Git-Sha":    []string{"yeet"},
		"Go-Version": []string{"1.14.1"},
	}
	SetVersionHeaders(w)
	gotHeader := w.Result().Header
	assert.Equal(t, wantHeader, gotHeader)
}

func TestJSON(t *testing.T) {
	version = Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	}
	wantJSON := `{"build_time":"2022-01-08T02:50:01Z","git_sha":"yeet","go_version":"1.14.1"}`
	gotJSON := string(JSON())
	if expected, got := wantJSON, gotJSON; expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}

func TestSet(t *testing.T) {
	version = Version{} // reset
	Set(Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	})

	wantVersion := Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	}
	gotVersion := version

	if expected, got := fmt.Sprintf("%#v", wantVersion), fmt.Sprintf("%#v", gotVersion); expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}

func TestString(t *testing.T) {
	version = Version{
		BuildTime: time.Unix(1641610201, 0).UTC(),
		GitSha:    "yeet",
		GoVersion: "1.14.1",
	}
	wantString := "yeet"
	gotString := String()

	if expected, got := wantString, gotString; expected != got {
		t.Errorf("\nwant: `%v`\n got: `%v`", expected, got)
	}
}
