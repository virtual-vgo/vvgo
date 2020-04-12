package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func Test_readUpload(t *testing.T) {
	testData := filepath.Join("..", "..", "pkg", "api", "testdata")

	type args struct {
		input    string
		project  string
		fileName string
	}

	type wants struct {
		output string
		upload api.Upload
		ok     bool
	}

	sheetBytes, err := ioutil.ReadFile(filepath.Join(testData, "sheet-music.pdf"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed %v", err)
	}
	clickBytes, err := ioutil.ReadFile(filepath.Join(testData, "click-track.mp3"))
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed %v", err)
	}

	for _, tt := range []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "file does not exist",
			args: args{
				input:    "y\n1, 2\ntrumpet , baritone\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "dne"),
			},
			wants: wants{
				ok: false,
			},
		},
		{
			name: "invalid media type",
			args: args{
				input:    "y\n1, 2\ntrumpet , baritone\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "invalid.data"),
			},
			wants: wants{
				ok: false,
			},
		},
		{
			name: "click/success",
			args: args{
				input:    "y\n1, 2\ntrumpet , baritone\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "click-track.mp3"),
			},
			wants: wants{
				ok:     true,
				output: ":: this is a click track [Y/n]? :: please enter part numbers (ex 1, 2): :: please enter part names (ex trumpet, flute): ",
				upload: api.Upload{
					UploadType:  api.UploadTypeClix,
					PartNames:   []string{"trumpet", "baritone"},
					PartNumbers: []uint8{1, 2},
					Project:     "01-snake-eater",
					FileName:    filepath.Join(testData, "click-track.mp3"),
					FileBytes:   clickBytes,
					ContentType: "audio/mpeg",
				},
			},
		},
		{
			name: "sheet/success",
			args: args{
				input:    "y\n1, 2\ntrumpet , baritone\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "sheet-music.pdf"),
			},
			wants: wants{
				ok:     true,
				output: ":: this is a music sheet [Y/n]? :: please enter part numbers (ex 1, 2): :: please enter part names (ex trumpet, flute): ",
				upload: api.Upload{
					UploadType:  api.UploadTypeSheets,
					PartNames:   []string{"trumpet", "baritone"},
					PartNumbers: []uint8{1, 2},
					Project:     "01-snake-eater",
					FileName:    filepath.Join(testData, "sheet-music.pdf"),
					FileBytes:   sheetBytes,
					ContentType: "application/pdf",
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotUpload api.Upload
			reader := bufio.NewReader(strings.NewReader(tt.args.input))
			writer := bytes.NewBuffer(nil)
			gotOk := readUpload(writer, reader, &gotUpload, tt.args.project, tt.args.fileName)
			assert.Equal(t, tt.wants.output, writer.String(), "output")
			if assert.Equal(t, tt.wants.ok, gotOk, "ok") && gotOk {
				compareUploads(t, &tt.wants.upload, &gotUpload)
			}
		})
	}
}

func compareUploads(t *testing.T, want, got *api.Upload) {
	if want == nil {
		assert.Nil(t, got, "upload")
		return
	}

	if assert.NotNil(t, got, "upload") {
		type flatUpload struct {
			UploadType   string
			PartNames    string
			PartNumbers  string
			Project      string
			FileName     string
			FileBytesSum string
			ContentType  string
		}

		flatten := func(upload *api.Upload) *flatUpload {
			return &flatUpload{
				UploadType:   upload.UploadType.String(),
				PartNames:    strings.Join(upload.PartNames, ","),
				PartNumbers:  fmt.Sprintf("%#v", upload.PartNumbers),
				Project:      upload.Project,
				FileName:     upload.FileName,
				FileBytesSum: fmt.Sprintf("%x", md5.Sum(upload.FileBytes)),
				ContentType:  upload.ContentType,
			}
		}

		assert.Equal(t, flatten(want), flatten(got), "upload")
	}
}

func Test_yesNo(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "EOF",
			input: "",
			want:  false,
		},
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

func Test_readPartNumbers(t *testing.T) {
	type wants struct {
		output  string
		numbers []uint8
	}
	for _, tt := range []struct {
		name  string
		input string
		wants wants
	}{
		{
			name:  "1",
			input: "1\n",
			wants: wants{
				output:  ":: please enter part numbers (ex 1, 2): ",
				numbers: []uint8{1},
			},
		},
		{
			name:  "1, 2",
			input: "1, 2\n",
			wants: wants{
				output:  ":: please enter part numbers (ex 1, 2): ",
				numbers: []uint8{1, 2},
			},
		},
		{
			name:  "invalid number",
			input: "cheese\n",
			wants: wants{
				output:  ":: please enter part numbers (ex 1, 2): ",
				numbers: nil,
			},
		},
		{
			name:  "EOF",
			input: "",
			wants: wants{
				output:  ":: please enter part numbers (ex 1, 2): ",
				numbers: nil,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var writer bytes.Buffer
			reader := bufio.NewReader(strings.NewReader(tt.input))
			gotNumbers := readPartNumbers(&writer, reader)
			assert.Equal(t, tt.wants.output, writer.String(), "output")
			assert.Equal(t, tt.wants.numbers, gotNumbers, "numbers")
		})
	}
}

func Test_readPartNames(t *testing.T) {
	type wants struct {
		output string
		names  []string
	}
	for _, tt := range []struct {
		name  string
		input string
		wants wants
	}{
		{
			name:  "trumpet",
			input: "trumpet\n",
			wants: wants{
				output: ":: please enter part names (ex trumpet, flute): ",
				names:  []string{"trumpet"},
			},
		},
		{
			name:  "trumpet, baritone",
			input: "trumpet, baritone\n",
			wants: wants{
				output: ":: please enter part names (ex trumpet, flute): ",
				names:  []string{"trumpet", "baritone"},
			},
		},
		{
			name:  "invalid instrument name",
			input: "not-an-instrument\n",
			wants: wants{
				output: ":: please enter part names (ex trumpet, flute): ",
				names:  nil,
			},
		},
		{
			name:  "EOF",
			input: "",
			wants: wants{
				output: ":: please enter part names (ex trumpet, flute): ",
				names:  nil,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var writer bytes.Buffer
			reader := bufio.NewReader(strings.NewReader(tt.input))
			gotNames := readPartNames(&writer, reader)
			assert.Equal(t, tt.wants.output, writer.String(), "output")
			assert.Equal(t, tt.wants.names, gotNames, "names")
		})
	}
}
