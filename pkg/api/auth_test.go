package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApiServer_Authenticate(t *testing.T) {
	type wants struct {
		code   int
		body   string
		header map[string]string
	}

	var newAuthRequest = func(url, user, pass string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, url, strings.NewReader(""))
		req.SetBasicAuth(user, pass)
		return req
	}

	for _, tt := range []struct {
		name    string
		config  Config
		request *http.Request
		wants   wants
	}{
		{
			name:    "success",
			request: newAuthRequest("/", "jackson", "the-earth-is-flat"),
			config:  Config{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants:   wants{code: http.StatusOK},
		},
		{
			name:    "incorrect user",
			request: newAuthRequest("/", "", "the-earth-is-flat"),
			config:  Config{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
		{
			name:    "incorrect pass",
			request: newAuthRequest("/", "jackson", ""),
			config:  Config{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
		{
			name:    "no auth",
			request: httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			config:  Config{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server := NewServer(MockObjectStore{}, tt.config)
			server.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				// do nothing
			})(recorder, tt.request)

			gotCode := recorder.Code
			gotBody := strings.TrimSpace(recorder.Body.String())

			if expected, got := tt.wants.code, gotCode; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, gotBody; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}

			for wantK := range tt.wants.header {
				fmt.Println(tt.wants.header)
				if expected, got := tt.wants.header[wantK], recorder.Header().Get(wantK); expected != got {
					t.Errorf("expected `%s: %v`, got `%s: %v`", wantK, expected, wantK, got)
				}
			}
		})
	}
}
