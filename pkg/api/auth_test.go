package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTokenAuth_Authenticate(t *testing.T) {
	var newRequest = func(url string, headers map[string]string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, url, strings.NewReader(""))
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req
	}
	for _, tt := range []struct {
		name      string
		request   *http.Request
		tokenAuth TokenAuth
		wantCode  int
	}{
		{
			name:      "success",
			request:   newRequest("/", map[string]string{"Authorization": "Bearer 196ddf804c7666d4-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"}),
			tokenAuth: TokenAuth{"196ddf804c7666d4-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"},
			wantCode:  http.StatusOK,
		},
		{
			name:      "empty map",
			request:   newRequest("/", map[string]string{"Virtual-VGO-Api-Secret": "Bearer 196ddf804c7666d4-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"}),
			tokenAuth: TokenAuth{},
			wantCode:  http.StatusUnauthorized,
		},
		{
			name:      "no token",
			request:   newRequest("/", map[string]string{}),
			tokenAuth: TokenAuth{"196ddf804c7666d4-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"},
			wantCode:  http.StatusUnauthorized,
		},
		{
			name:      "incorrect token",
			request:   newRequest("/", map[string]string{"Virtual-VGO-Api-Secret": "Bearer 8d32ff4a91a530bc-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"}),
			tokenAuth: TokenAuth{"196ddf804c7666d4-8d32ff4a91a530bc-c5c7cde4a26096ad-67758135226bfb2e"},
			wantCode:  http.StatusUnauthorized,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			tt.tokenAuth.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// do nothing
			})).ServeHTTP(recorder, tt.request)

			gotCode := recorder.Code

			if expected, got := tt.wantCode, gotCode; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func TestBasicAuth_Authenticate(t *testing.T) {
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
		config  ServerConfig
		request *http.Request
		wants   wants
	}{
		{
			name:    "success",
			request: newAuthRequest("/", "jackson", "the-earth-is-flat"),
			config:  ServerConfig{MemberUser: "jackson", MemberPass: "the-earth-is-flat"},
			wants:   wants{code: http.StatusOK},
		},
		{
			name:    "incorrect user",
			request: newAuthRequest("/", "", "the-earth-is-flat"),
			config:  ServerConfig{MemberUser: "jackson", MemberPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
		{
			name:    "incorrect pass",
			request: newAuthRequest("/", "jackson", ""),
			config:  ServerConfig{MemberUser: "jackson", MemberPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
		{
			name:    "no auth",
			request: httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			config:  ServerConfig{MemberUser: "jackson", MemberPass: "the-earth-is-flat"},
			wants: wants{
				code:   http.StatusUnauthorized,
				body:   "authorization failed",
				header: map[string]string{"WWW-Authenticate": `Basic charset="UTF-8"`},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server := BasicAuth{tt.config.MemberUser: tt.config.MemberPass}
			server.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// do nothing
			})).ServeHTTP(recorder, tt.request)

			gotCode := recorder.Code
			gotBody := strings.TrimSpace(recorder.Body.String())

			if expected, got := tt.wants.code, gotCode; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, gotBody; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}

			for wantK := range tt.wants.header {
				if expected, got := tt.wants.header[wantK], recorder.Header().Get(wantK); expected != got {
					t.Errorf("expected `%s: %v`, got `%s: %v`", wantK, expected, wantK, got)
				}
			}
		})
	}
}
