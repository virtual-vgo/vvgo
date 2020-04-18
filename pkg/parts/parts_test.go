package parts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"testing"
	"time"
)

type MockBucket struct {
	putObject func(ctx context.Context, name string, object *storage.Object) bool
	getObject func(ctx context.Context, name string, dest *storage.Object) bool
}

func (x *MockBucket) PutObject(ctx context.Context,name string, object *storage.Object) bool {
	return x.putObject(ctx, name, object)
}

func (x *MockBucket) GetObject(ctx context.Context,name string, dest *storage.Object) bool {
	return x.getObject(ctx, name, dest)
}

type MockLocker struct {
	lock   func(ctx context.Context) bool
	unlock func(ctx context.Context)
}

func (x *MockLocker) Lock(ctx context.Context) bool {
	return x.lock(ctx)
}

func (x *MockLocker) Unlock(ctx context.Context,) {
	x.unlock(ctx)
}

func TestParts_Init(t *testing.T) {
	wantName := DataFile
	wantObject := &storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[]`)),
	}

	var gotName string
	var gotObject *storage.Object

	parts := Parts{Bucket: &MockBucket{
		putObject: func(ctx context.Context, name string, object *storage.Object) bool {
			gotName = name
			gotObject = object
			return true
		}},
	}

	parts.Init(context.Background())
	assert.Equal(t, wantName, gotName, "name")
	assert.Equal(t, gotObject, wantObject, "object")
}

func TestParts_List(t *testing.T) {
	wantName := DataFile
	wantList := []Part{{ID: ID{
		Project: "cheese",
		Name:    "broccoli",
		Number:  3,
	}}}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&wantList); err != nil {
		t.Fatalf("json.Encode() failed: %v", err)
	}

	var gotName string
	parts := Parts{Bucket: &MockBucket{
		getObject: func(ctx context.Context, name string, object *storage.Object) bool {
			gotName = name
			*object = storage.Object{
				ContentType: "application/json",
				Buffer:      buffer,
			}
			return true
		}},
	}
	gotList := parts.List(context.Background())

	assert.Equal(t, wantName, gotName, "name")
	assert.Equal(t, wantList, gotList, "object")
}

func TestParts_Save(t *testing.T) {
	type args struct {
		parts     []Part
		fileBytes []byte
	}

	cmdArgs := args{
		parts: []Part{{ID: ID{
			Project: "01-snake-eater",
			Name:    "trumpet",
			Number:  3,
		}}},
	}

	wantOk := true
	wantNames := []string{
		fmt.Sprintf("%s-%s", DataFile, time.Now().UTC().Format(time.RFC3339)),
		DataFile,
	}
	wantObjects := []storage.Object{
		{ContentType: "application/json", Buffer: *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5},{"project":"01-snake-eater","part_name":"trumpet","part_number":3}]`))},
		{ContentType: "application/json", Buffer: *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5},{"project":"01-snake-eater","part_name":"trumpet","part_number":3}]`))},
	}

	var gotNames []string
	var gotObjects []storage.Object
	parts := Parts{
		Bucket: &MockBucket{
			getObject: func(ctx context.Context, name string, object *storage.Object) bool {
				*object = storage.Object{
					ContentType: "application/json",
					Buffer:      *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5}]`))}
				return true
			},
			putObject: func(ctx context.Context, name string, object *storage.Object) bool {
				gotNames = append(gotNames, name)
				gotObjects = append(gotObjects, *object)
				return true
			},
		},
		Locker: &MockLocker{
			lock:   func(ctx context.Context) bool { return true },
			unlock: func(ctx context.Context) {},
		},
	}
	gotOk := parts.Save(nil, cmdArgs.parts)

	assert.Equal(t, wantOk, gotOk, "ok")
	assert.Equal(t, wantNames, gotNames, "names")
	if want, got := objectsToString(wantObjects), objectsToString(gotObjects); want != got {
		t.Errorf("\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func objectsToString(objects []storage.Object) string {
	var str string
	for _, object := range objects {
		str += fmt.Sprintf("content-type: `%s`, body: `%s`\n", object.ContentType, strings.TrimSpace(object.Buffer.String()))
	}
	return strings.TrimSpace(str)
}

func TestPart_String(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "danish", Number: 1},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "Project: cheese Part: Danish #1"
	assert.Equal(t, want, part.String())
}

func TestPart_SheetLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "danish", Number: 1},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "/download?bucket=sheets&object=sheet.pdf"
	assert.Equal(t, want, part.SheetLink("sheets"))
}

func TestPart_ClickLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "danish", Number: 1},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "/download?bucket=clix&object=click.mp3"
	assert.Equal(t, want, part.ClickLink("clix"))
}

func TestPart_Validate(t *testing.T) {
	type fields struct {
		Project    string
		PartName   string
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
				Project:    "01-snake-eater",
				PartName:   "trumpet",
				PartNumber: 6,
			},
			want: nil,
		},
		{
			name: "missing project",
			fields: fields{
				PartName:   "trumpet",
				PartNumber: 6,
			},
			want: projects.ErrNotFound,
		},
		{
			name: "missing part name",
			fields: fields{
				Project:    "01-snake-eater",
				PartNumber: 6,
			},
			want: ErrInvalidPartName,
		},
		{
			name: "invalid part name",
			fields: fields{
				Project:    "01-snake-eater",
				PartName:   "not-an-instrument",
				PartNumber: 6,
			},
			want: ErrInvalidPartName,
		},
		{
			name: "missing part number",
			fields: fields{
				Project:  "01-snake-eater",
				PartName: "trumpet",
			},
			want: ErrInvalidPartNumber,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := &Part{
				ID: ID{
					Project: tt.fields.Project,
					Name:    tt.fields.PartName,
					Number:  tt.fields.PartNumber,
				},
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func Test_mergeChanges(t *testing.T) {
	type args struct {
		src     []Part
		changes []Part
	}
	for _, tt := range []struct {
		name string
		args args
		want []Part
	}{
		{
			name: "",
			args: args{
				src: []Part{
					{
						ID:   ID{Project: "cheese", Name: "danish", Number: 1},
						Clix: []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
					},
					{
						ID:     ID{Project: "turkey", Name: "club", Number: 3},
						Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
						Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
					},
				},
				changes: []Part{
					{
						ID:     ID{Project: "cheese", Name: "danish", Number: 1},
						Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
						Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
					},
					{
						ID:     ID{Project: "waffle", Name: "cone", Number: 2},
						Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
						Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
					},
				},
			},
			want: []Part{
				{
					ID: ID{Project: "cheese", Name: "danish", Number: 1},
					Clix: []Link{
						{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)},
						{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)},
					},
					Sheets: []Link{
						{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)},
					},
				},
				{
					ID:     ID{Project: "turkey", Name: "club", Number: 3},
					Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
					Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
				},
				{
					ID:     ID{Project: "waffle", Name: "cone", Number: 2},
					Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
					Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeChanges(tt.args.src, tt.args.changes)
			if want, got := fmt.Sprintf("%#v", tt.want), fmt.Sprintf("%#v", got); want != got {
				t.Errorf("\nwant: %s\n got: %s", want, got)
			}
		})
	}
}
