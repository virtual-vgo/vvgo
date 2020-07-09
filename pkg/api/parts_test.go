package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"strings"
	"testing"
	"time"
)

func Test_partsToCSV(t *testing.T) {
	assert.Equal(t, `01-snake-eater,trumpet 1,,New-click.mp3,1
`, string(partsToCSV([]parts.Part{{
		ID:   parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Meta: parts.Meta{ScoreOrder: 1},
		Clix: []parts.Link{{ObjectKey: "New-click.mp3", CreatedAt: time.Unix(2, 0)},
			{ObjectKey: "Old-click.mp3", CreatedAt: time.Unix(1, 0)}},
		Sheets: []parts.Link{},
	}})))
}

func Test_partsFromCSV(t *testing.T) {
	assert.Equal(t, []parts.Part{{
		ID:   parts.ID{Project: "01-snake-eater", Name: "trumpet 1"},
		Meta: parts.Meta{ScoreOrder: 1},
		Clix: []parts.Link{},
		Sheets: []parts.Link{},
	}}, partsFromCSV(strings.NewReader(`01-snake-eater,trumpet 1,,,1
`)))
}
