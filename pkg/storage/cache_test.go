package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestMemHash_Keys(t *testing.T) {
	ctx := context.Background()
	hash := MemHash{Map: map[string][]byte{
		"I'd Say She Found A Way Out": []byte("Wouldn't You?"),
		"Dying Doesn't Scare Me.":     []byte("I Went To Work Every Day Prepared To Die In A Tiger Cage."),
	}}
	wantKeys := []string{"I'd Say She Found A Way Out", "Dying Doesn't Scare Me."}
	gotKeys, err := hash.HKeys(ctx)
	assert.NoError(t, err)
	sort.Strings(gotKeys)
	sort.Strings(wantKeys)
	assert.Equal(t, wantKeys, gotKeys)
}

func TestMemHash_Set(t *testing.T) {
	ctx := context.Background()
	hash := MemHash{}
	assert.NoError(t, hash.HSet(ctx, "Dying Doesn't Scare Me.",
		MarshalString("I Went To Work Every Day Prepared To Die In A Tiger Cage.")))
	wantMap := map[string][]byte{
		"Dying Doesn't Scare Me.": []byte("I Went To Work Every Day Prepared To Die In A Tiger Cage."),
	}
	assert.Equal(t, wantMap, hash.Map)
}

func TestMemHash_Get(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		hash := MemHash{Map: map[string][]byte{
			"Dying Doesn't Scare Me.": []byte("I Went To Work Every Day Prepared To Die In A Tiger Cage."),
		}}
		var gotString string
		assert.NoError(t, hash.HGet(ctx, "Dying Doesn't Scare Me.", (*UnmarshalString)(&gotString)))
		assert.Equal(t, "I Went To Work Every Day Prepared To Die In A Tiger Cage.", gotString)
	})
	t.Run("doesnt exist", func(t *testing.T) {
		ctx := context.Background()
		hash := MemHash{Map: map[string][]byte{
			"Dying Doesn't Scare Me.": []byte("I Went To Work Every Day Prepared To Die In A Tiger Cage."),
		}}
		var gotString string
		assert.Equal(t, ErrKeyIsEmpty, hash.HGet(ctx, "I'd Say She Found A Way Out", (*UnmarshalString)(&gotString)))
	})

}
