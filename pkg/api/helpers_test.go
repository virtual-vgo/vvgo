package api

import (
	"net/http"
	"testing"
)

func Test_acceptsType(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header http.Header
		arg    string
		want   bool
	}{
		{
			name: "yep",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg:  "application/xml",
			want: true,
		},
		{
			name: "nope",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg: "sour/cream", want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if expected, got := tt.want, acceptsType(&http.Request{Header: tt.header}, tt.arg); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}
