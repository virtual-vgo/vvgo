package parts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"testing"
	"time"
)

func TestParts_Init(t *testing.T) {
	wantName := DataFile
	wantObject := &storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[]`)),
	}

	var gotName string
	var gotObject *storage.Object

	bucket, err := storage.NewBucket(context.Background(), "test")
	require.NoError(t, err, "storage.NewBucket()")
	parts := Parts{Bucket: bucket}
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
	bucket, err := storage.NewBucket(context.Background(), "test")
	require.NoError(t, err, "storage.NewBucket()")
	parts := Parts{Bucket: bucket}
	gotList,_ := parts.List(context.Background())

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
	bucket, err := storage.NewBucket(context.Background(), "test")
	require.NoError(t, err, "storage.NewBucket()")
	parts := Parts{
		Bucket: bucket,
		Locker: locker.NewLocker("test"),
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
				PartName:   "",
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
