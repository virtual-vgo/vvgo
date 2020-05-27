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
				input:    "trumpet , baritone\n1, 2\n",
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
				input:    "trumpet , baritone\n1, 2\n",
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
				input:    "trumpet 1, trumpet 2,  baritone 1,baritone 2\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "click-track.mp3"),
			},
			wants: wants{
				ok:     true,
				output: ":: upload type: clix\n:: leave empty to skip | Ctrl+C to quit\n:: part names (ex. trumpet, flute): ",
				upload: api.Upload{
					UploadType:  api.UploadTypeClix,
					PartNames:   []string{"trumpet 1", "trumpet 2", "baritone 1", "baritone 2"},
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
				input:    "trumpet 1, trumpet 2,  baritone 1,baritone 2\n",
				project:  "01-snake-eater",
				fileName: filepath.Join(testData, "sheet-music.pdf"),
			},
			wants: wants{
				ok:     true,
				output: ":: upload type: sheets\n:: leave empty to skip | Ctrl+C to quit\n:: part names (ex. trumpet, flute): ",
				upload: api.Upload{
					UploadType:  api.UploadTypeSheets,
					PartNames:   []string{"trumpet 1", "trumpet 2", "baritone 1", "baritone 2"},
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
			gotError := readUpload(writer, reader, &gotUpload, tt.args.project, tt.args.fileName)
			assert.Equal(t, tt.wants.output, writer.String(), "output")
			if tt.wants.ok && assert.NoError(t, gotError, "error") {
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
			Project      string
			FileName     string
			FileBytesSum string
			ContentType  string
		}

		flatten := func(upload *api.Upload) *flatUpload {
			return &flatUpload{
				UploadType:   upload.UploadType.String(),
				PartNames:    strings.Join(upload.PartNames, ","),
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
				output: ":: part names (ex. trumpet, flute): ",
				names:  []string{"trumpet"},
			},
		},
		{
			name:  "trumpet, baritone",
			input: "trumpet, baritone\n",
			wants: wants{
				output: ":: part names (ex. trumpet, flute): ",
				names:  []string{"trumpet", "baritone"},
			},
		},
		{
			name:  "EOF",
			input: "",
			wants: wants{
				output: ":: part names (ex. trumpet, flute): ",
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
