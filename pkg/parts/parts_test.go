package parts

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"sort"
	"testing"
	"time"
)

func TestParts_List(t *testing.T) {
	ctx := context.Background()
	parts := Parts{
		Hash:   new(storage.MemHash),
		Locker: &storage.MemLocker{},
	}
	wantList := []Part{{ID: ID{
		Project: "cheese",
		Name:    "broccoli",
		Number:  3,
	}}}

	require.NoError(t, parts.Hash.HSet(ctx, DataKey, &Part{ID: ID{
		Project: "cheese",
		Name:    "broccoli",
		Number:  3,
	}}), "cache.PutObject()")

	gotList, err := parts.List(context.Background())
	assert.NoError(t, err, "parts.List()")
	assert.Equal(t, wantList, gotList, "object")
}

func TestParts_Save(t *testing.T) {
	ctx := context.Background()
	parts := Parts{
		Hash:   new(storage.MemHash),
		Locker: &storage.MemLocker{},
	}

	// load some dummy data into the cache

	require.NoError(t, parts.Hash.HSet(ctx, "01-snake-eater-trumpet-1", &Part{
		ID:   ID{Project: "01-snake-eater", Name: "trumpet", Number: 1},
		Clix: []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
	}), "Hash.HSet")

	require.NoError(t, parts.Hash.HSet(ctx, "01-snake-eater-accordion-3", &Part{
		ID:     ID{Project: "01-snake-eater", Name: "accordion", Number: 3},
		Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
		Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
	}), "Hash.HSet")

	// now save some data

	require.NoError(t, parts.Save(ctx, []Part{
		{
			ID:     ID{Project: "01-snake-eater", Name: "trumpet", Number: 1},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "triangle", Number: 2},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
	}), "parts.Save()")

	wantParts := []Part{
		{
			ID: ID{Project: "01-snake-eater", Name: "trumpet", Number: 1},
			Clix: []Link{
				{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)},
				{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)},
			},
			Sheets: []Link{
				{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)},
			},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "accordion", Number: 3},
			Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
			Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "triangle", Number: 2},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
	}
	gotParts, err := parts.List(ctx)
	assert.NoError(t, err, "parts.List()")
	SortParts(gotParts).Sort()
	SortParts(wantParts).Sort()
	assert.Equal(t, wantParts, gotParts)
}

type SortParts []Part

func (x SortParts) Sort()              { sort.Sort(x) }
func (x SortParts) Len() int           { return len(x) }
func (x SortParts) Less(i, j int) bool { return x[i].ID.String() < x[j].ID.String() }
func (x SortParts) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func TestPart_String(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet", Number: 1},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "Project: cheese Part: Trumpet #1"
	assert.Equal(t, want, part.String())
}

func TestPart_SheetLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet", Number: 1},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "/download?bucket=sheets&object=sheet.pdf"
	assert.Equal(t, want, part.SheetLink("sheets"))
}

func TestPart_ClickLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet", Number: 1},
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
