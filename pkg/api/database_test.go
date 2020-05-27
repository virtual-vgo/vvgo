package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"testing"
	"time"
)

func TestDatabase_Backup(t *testing.T) {
	ctx := context.Background()
	database := &Database{
		Parts:  newParts(),
		Distro: newBucket(t),
	}
	require.NoError(t, database.Parts.Save(ctx, []parts.Part{{
		ID:   parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix: []parts.Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0).UTC()}},
	}}))
	got, err := database.Backup(ctx)
	require.NoError(t, err, "database.Backup()")
	assert.NotZero(t, got.Timestamp)
	assert.Equal(t, []parts.Part{{
		ID:     parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Sheets: []parts.Link{},
		Clix:   []parts.Link{{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0).UTC()}},
	}}, got.Parts)
	assert.Equal(t, string(version.JSON()), string(got.ApiVersion))
}

func TestDatabase_Restore(t *testing.T) {
	ctx := context.Background()
	database := &Database{
		Parts:  newParts(),
		Distro: newBucket(t),
	}
	require.NoError(t, database.Parts.Save(ctx, []parts.Part{{
		ID:     parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix:   []parts.Link{{ObjectKey: "OLD-click.mp3", CreatedAt: time.Unix(1, 0).UTC()}},
		Sheets: []parts.Link{{ObjectKey: "OLD-sheet.pdf", CreatedAt: time.Unix(1, 0).UTC()}},
	}}), "parts.Save()")

	require.NoError(t, database.Restore(ctx, DatabaseBackup{Parts: []parts.Part{{
		ID:     parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix:   []parts.Link{{ObjectKey: "NEW-click.mp3", CreatedAt: time.Unix(2, 0).UTC()}},
		Sheets: []parts.Link{{ObjectKey: "NEW-sheet.pdf", CreatedAt: time.Unix(2, 0).UTC()}},
	}}}))

	gotList, err := database.Parts.List(ctx)
	assert.NoError(t, err, "parts.List()")
	assert.Equal(t, []parts.Part{{
		ID:     parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Clix:   []parts.Link{{ObjectKey: "NEW-click.mp3", CreatedAt: time.Unix(2, 0).UTC()}},
		Sheets: []parts.Link{{ObjectKey: "NEW-sheet.pdf", CreatedAt: time.Unix(2, 0).UTC()}},
	}}, gotList)
}
