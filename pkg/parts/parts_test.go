package parts

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	var redisConfig redis.Config
	envconfig.MustProcess("REDIS", &redisConfig)
	redis.Initialize(redisConfig)
}

func newParts() RedisParts {
	return RedisParts{namespace: "testing" + strconv.Itoa(lrand.Int())}
}

func TestParts_List(t *testing.T) {
	ctx := context.Background()
	parts := newParts()

	wantList := []Part{{
		ID: ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix: []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)},
			{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
		Sheets: []Link{},
	}}

	require.NoError(t, parts.Save(ctx, []Part{{
		ID: ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix: []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)},
			{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
	}}))

	gotList, err := parts.List(context.Background())
	assert.NoError(t, err, "parts.List()")
	assertEqualParts(t, wantList, gotList)
}

func TestRedisParts_DeleteAll(t *testing.T) {
	ctx := context.Background()
	parts := newParts()

	require.NoError(t, parts.Save(ctx, []Part{{
		ID: ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix: []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)},
			{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
	}}))

	parts.DeleteAll(ctx)
	gotList, err := parts.List(context.Background())
	assert.NoError(t, err, "parts.List()")
	assert.Empty(t, gotList)
}

func TestParts_Save(t *testing.T) {
	ctx := context.Background()
	parts := newParts()

	require.NoError(t, parts.Save(ctx, []Part{
		{
			ID:   ID{Project: "01-snake-eater", Name: "trumpet 1"},
			Clix: []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "accordion 3"},
			Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
			Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
		},
	}))

	// now save some merging changes

	require.NoError(t, parts.Save(ctx, []Part{
		{
			ID:     ID{Project: "01-snake-eater", Name: "trumpet 1"},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "triangle 2"},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
	}), "parts.Save()")

	wantParts := []Part{
		{
			ID: ID{Project: "01-snake-eater", Name: "trumpet 1"},
			Clix: []Link{
				{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)},
				{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)},
			},
			Sheets: []Link{
				{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)},
			},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "accordion 3"},
			Clix:   []Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
			Sheets: []Link{{ObjectKey: "Old-sheet.pdf", CreatedAt: time.Unix(1, 0)}},
		},
		{
			ID:     ID{Project: "01-snake-eater", Name: "triangle 2"},
			Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
			Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
		},
	}
	gotParts, err := parts.List(ctx)
	assert.NoError(t, err, "parts.List()")
	assertEqualParts(t, wantParts, gotParts)
}

type SortParts []Part

func (x SortParts) Sort()              { sort.Sort(x) }
func (x SortParts) Len() int           { return len(x) }
func (x SortParts) Less(i, j int) bool { return x[i].RedisKey() < x[j].RedisKey() }
func (x SortParts) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func assertEqualParts(t *testing.T, want []Part, got []Part) {
	SortParts(want).Sort()
	SortParts(got).Sort()

	if len(want) != len(got) {
		assert.Equal(t, want, got)
	}
	for i := range want {
		assert.Equal(t, want[i].RedisKey(), got[i].RedisKey(), "part.ID")
		assertEqualLinks(t, want[i].Sheets, got[i].Sheets, "part.Sheets")
		assertEqualLinks(t, want[i].Clix, got[i].Clix, "part.Clix")

	}
}

func assertEqualLinks(t *testing.T, want []Link, got []Link, pre string) {
	if len(want) != len(got) {
		assert.Equal(t, want, got)
	}
	for i := range want {
		assert.Equal(t, want[i].ObjectKey, got[i].ObjectKey, pre+".ObjectKey")
		assert.Equal(t, want[i].CreatedAt.UTC().String(), got[i].CreatedAt.UTC().String(), pre+".CreatedAt")
	}
}

func TestPart_String(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet 1"},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "Project: cheese Part: Trumpet 1"
	assert.Equal(t, want, part.String())
}

func TestPart_SheetLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet 1"},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "/download?bucket=sheets&object=sheet.pdf"
	assert.Equal(t, want, part.SheetLink("sheets"))
}

func TestPart_ClickLink(t *testing.T) {
	part := Part{
		ID:     ID{Project: "cheese", Name: "trumpet 1"},
		Clix:   []Link{{"click.mp3", time.Now()}},
		Sheets: []Link{{"sheet.pdf", time.Now()}},
	}
	want := "/download?bucket=clix&object=click.mp3"
	assert.Equal(t, want, part.ClickLink("clix"))
}

func TestPart_Validate(t *testing.T) {
	type fields struct {
		Project  string
		PartName string
	}
	for _, tt := range []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "valid",
			fields: fields{
				Project:  "01-snake-eater",
				PartName: "trumpet 6",
			},
			want: nil,
		},
		{
			name: "missing project",
			fields: fields{
				PartName: "trumpet 6",
			},
			want: projects.ErrNotFound,
		},
		{
			name: "missing part name",
			fields: fields{
				Project: "01-snake-eater",
			},
			want: ErrInvalidPartName,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := &Part{
				ID: ID{
					Project: tt.fields.Project,
					Name:    tt.fields.PartName,
				},
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func TestPart_ZScore(t *testing.T) {
	part := Part{
		ID:     ID{Project: "01-snake-eater", Name: "triangle 2"},
		Clix:   []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}},
		Sheets: []Link{{ObjectKey: "New-sheet.pdf", CreatedAt: time.Unix(2, 0)}},
	}
	assert.Equal(t, 1, part.ZScore())
}

func TestPart_RedisKey(t *testing.T) {
	part := Part{ID: ID{Project: "01-snake-eater", Name: "triangle 2"}}
	assert.Equal(t, "01-snake-eater:triangle 2", part.RedisKey())
}

func TestPart_DecodeRedisKey(t *testing.T) {
	var got Part
	got.DecodeRedisKey("01-snake-eater:triangle 2")
	assert.Equal(t, Part{ID: ID{Project: "01-snake-eater", Name: "triangle 2"}}, got)
}

func TestLink_DecodeRedisString(t *testing.T) {
	linkString := `{"object_key":"New-click.mp3","created_at":"1969-12-31T16:00:02-08:00"}`
	var got Link
	got.DecodeRedisString(linkString)
	assertEqualLinks(t, []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}}, []Link{got}, "link")
}

func TestLink_EncodeRedisString(t *testing.T) {
	link := Link{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}
	var got Link
	got.DecodeRedisString(link.EncodeRedisString())
	assertEqualLinks(t, []Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}}, []Link{got}, "link")
}

func TestLink_ZScore(t *testing.T) {
	link := Link{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)}
	assert.Equal(t, 2, link.ZScore())
}
