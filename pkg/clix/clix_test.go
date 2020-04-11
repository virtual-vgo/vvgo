package clix

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"testing"
	"time"
)

type MockBucket struct {
	putObject   func(name string, object *storage.Object) bool
	getObject   func(name string, dest *storage.Object) bool
	downloadURL func(name string) (string, error)
}

func (x *MockBucket) PutObject(name string, object *storage.Object) bool {
	return x.putObject(name, object)
}

func (x *MockBucket) GetObject(name string, dest *storage.Object) bool {
	return x.getObject(name, dest)
}

func (x *MockBucket) DownloadURL(name string) (string, error) {
	return x.downloadURL(name)
}

type MockLocker struct {
	lock   func(ctx context.Context) bool
	unlock func()
}

func (x *MockLocker) Lock(ctx context.Context) bool {
	return x.lock(ctx)
}

func (x *MockLocker) Unlock() {
	x.unlock()
}

func TestClix_Init(t *testing.T) {
	wantName := DataFile
	wantObject := &storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[]`)),
	}

	var gotName string
	var gotObject *storage.Object

	clix := Clix{Bucket: &MockBucket{
		putObject: func(name string, object *storage.Object) bool {
			gotName = name
			gotObject = object
			return true
		}},
	}

	clix.Init()
	assert.Equal(t, wantName, gotName, "name")
	assert.Equal(t, gotObject, wantObject, "object")
}

func TestClix_List(t *testing.T) {
	wantName := DataFile
	wantList := []Click{
		{
			Project:    "cheese",
			PartName:   "turnip",
			PartNumber: 5,
			FileKey:    "0xff",
		},
	}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&wantList); err != nil {
		t.Fatalf("json.Encode() failed: %v", err)
	}

	var gotName string
	clix := Clix{Bucket: &MockBucket{
		getObject: func(name string, object *storage.Object) bool {
			gotName = name
			*object = storage.Object{
				ContentType: "application/json",
				Buffer:      buffer,
			}
			return true
		}},
	}
	gotList := clix.List()

	assert.Equal(t, wantName, gotName, "name")
	assert.Equal(t, wantList, gotList, "object")
}

func TestClix_Store(t *testing.T) {
	type args struct {
		clix      []Click
		fileBytes []byte
		file      File
	}

	cmdArgs := args{
		clix: []Click{{
			Project:    "cheese",
			PartName:   "broccoli",
			PartNumber: 3,
		}},
		file: File{
			MediaType: "audio/mpeg",
			Ext:       ".mp3",
			Bytes:     []byte("pretend i'm an mpeg file ;)"),
		},
	}

	wantOk := true
	wantNames := []string{
		fmt.Sprintf("%x.mp3", md5.Sum([]byte("pretend i'm an mpeg file ;)"))),
		fmt.Sprintf("%s-%s", DataFile, time.Now().UTC().Format(time.RFC3339)),
		DataFile,
	}
	wantObjects := []storage.Object{
		{ContentType: "audio/mpeg", Buffer: *bytes.NewBuffer([]byte("pretend i'm an mpeg file ;)"))},
		{ContentType: "application/json", Buffer: *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5,"file_key":"0xff"},{"project":"cheese","part_name":"broccoli","part_number":3,"file_key":"9f442f3c87d9d78dc975cf34591ddcc0.mp3"}]`))},
		{ContentType: "application/json", Buffer: *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5,"file_key":"0xff"},{"project":"cheese","part_name":"broccoli","part_number":3,"file_key":"9f442f3c87d9d78dc975cf34591ddcc0.mp3"}]`))},
	}

	var gotNames []string
	var gotObjects []storage.Object
	mockClix := Clix{
		Bucket: &MockBucket{
			getObject: func(name string, object *storage.Object) bool {
				*object = *storage.NewJSONObject(bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5,"file_key":"0xff"}]`)))
				return true
			},
			putObject: func(name string, object *storage.Object) bool {
				gotNames = append(gotNames, name)
				gotObjects = append(gotObjects, *object)
				return true
			},
		},
		Locker: &MockLocker{
			lock:   func(ctx context.Context) bool { return true },
			unlock: func() {},
		},
	}

	gotOk := mockClix.Store(nil, cmdArgs.clix, &cmdArgs.file)

	assert.Equal(t, wantOk, gotOk, "ok")
	assert.Equal(t, wantNames, gotNames, "names")
	if want, got := objectsToString(wantObjects), objectsToString(gotObjects); want != got {
		t.Errorf("\nwant: %s\n got: %s", want, got)
	}
}

func objectsToString(objects []storage.Object) string {
	var str string
	for _, object := range objects {
		str += fmt.Sprintf("content-type: `%s`, body: `%s`, ", object.ContentType, strings.TrimSpace(object.Buffer.String()))
	}
	return strings.TrimSpace(str)
}

func TestClick_String(t *testing.T) {
	click := Click{FileKey: "mock-file-key"}
	want := fmt.Sprintf("Project: %s Part: %s-%d", click.Project, click.PartName, click.PartNumber)
	assert.Equal(t, want, click.String())
}

func TestClick_Link(t *testing.T) {
	click := Click{FileKey: "mock-file-key"}
	want := "/download?bucket=clix&object=mock-file-key"
	assert.Equal(t, want, click.Link("clix"))
}

func TestClick_ObjectKey(t *testing.T) {
	click := Click{FileKey: "mock-file-key"}
	want := "mock-file-key"
	assert.Equal(t, want, click.ObjectKey())
}

func TestClick_Validate(t *testing.T) {
	type fields struct {
		Project    string
		Instrument string
		PartNumber uint8
	}
	for _, tt := range []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "valid",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: nil,
		},
		{
			name: "missing project",
			fields: fields{
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: ErrMissingProject,
		},
		{
			name: "missing instrument",
			fields: fields{
				Project:    "test-project",
				PartNumber: 6,
			},
			want: ErrMissingPartName,
		},
		{
			name: "missing part number",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
			},
			want: ErrMissingPartNumber,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := &Click{
				Project:    tt.fields.Project,
				PartName:   tt.fields.Instrument,
				PartNumber: tt.fields.PartNumber,
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}
