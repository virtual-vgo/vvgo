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
	ctx := context.Background()
	parts := Parts{
		Cache:  storage.NewCache(storage.CacheOpts{}),
		Locker: locker.NewLocker(locker.Opts{}),
	}

	wantObject := storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[]`)),
	}

	require.NoError(t, parts.Init(ctx), "init")

	var gotObject storage.Object
	assert.NoError(t, parts.GetObject(context.Background(), DataFile, &gotObject))
	assert.Equal(t, objectToString(wantObject), objectToString(gotObject))
}

func TestParts_List(t *testing.T) {
	ctx := context.Background()
	parts := Parts{
		Cache:  storage.NewCache(storage.CacheOpts{}),
		Locker: locker.NewLocker(locker.Opts{}),
	}
	wantList := []Part{{ID: ID{
		Project: "cheese",
		Name:    "broccoli",
		Number:  3,
	}}}

	// load the cache with a dummy object
	obj := storage.Object{ContentType: "application/json"}
	require.NoError(t, json.NewEncoder(&obj.Buffer).Encode([]Part{{ID: ID{
		Project: "cheese",
		Name:    "broccoli",
		Number:  3,
	}}}), "json.Encode()")
	require.NoError(t, parts.Cache.PutObject(ctx, DataFile, &obj), "cache.PutObject()")

	gotList, err := parts.List(context.Background())
	assert.NoError(t, err, "parts.List()")
	assert.Equal(t, wantList, gotList, "object")
}

func TestParts_Save(t *testing.T) {
	ctx := context.Background()
	parts := Parts{
		Cache:  storage.NewCache(storage.CacheOpts{}),
		Locker: locker.NewLocker(locker.Opts{}),
	}
	require.NoError(t, parts.Init(ctx))

	// load the cache with a dummy object
	obj := storage.Object{ContentType: "application/json"}
	require.NoError(t, json.NewEncoder(&obj.Buffer).Encode([]Part{{ID: ID{
		Project: "cheese",
		Name:    "turnip",
		Number:  5,
	}}}), "json.Encode()")
	require.NoError(t, parts.Cache.PutObject(ctx, DataFile, &obj), "cache.PutObject()")

	type args struct {
		parts []Part
	}

	cmdArgs := args{
		parts: []Part{{ID: ID{
			Project: "01-snake-eater",
			Name:    "trumpet",
			Number:  3,
		}}},
	}

	wantObject := storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[{"project":"cheese","part_name":"turnip","part_number":5},{"project":"01-snake-eater","part_name":"trumpet","part_number":3}]`)),
	}

	assert.NoError(t, parts.Save(ctx, cmdArgs.parts), "parts.Save()")

	// check the data file
	var gotObject storage.Object
	assert.NoError(t, parts.GetObject(context.Background(), DataFile, &gotObject))
	assert.Equal(t, objectToString(wantObject), objectToString(gotObject))
}

func objectToString(object storage.Object) string {
	return fmt.Sprintf("content-type: `%s`, body: `%s`\n", object.ContentType, strings.TrimSpace(object.Buffer.String()))
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
