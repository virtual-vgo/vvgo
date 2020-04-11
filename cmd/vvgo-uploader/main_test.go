package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func Test_yesNo(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty",
			input: "\n",
			want:  true,
		},
		{
			name:  "y",
			input: "           y \n",
			want:  true,
		},
		{
			name:  "Y",
			input: " Y         \n",
			want:  true,
		},
		{
			name:  "yes",
			input: " yes         \n",
			want:  true,
		},
		{
			name:  "n",
			input: "   n \n",
			want:  false,
		},
		{
			name:  "N",
			input: "N       \n",
			want:  false,
		},
		{
			name:  "x",
			input: "x       \n",
			want:  false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var writer bytes.Buffer
			reader := bufio.NewReader(strings.NewReader(tt.input))
			if got := yesNo(&writer, reader, ""); got != tt.want {
				t.Errorf("yesNo() = %v, want %v", got, tt.want)
			}
		})
	}
}
