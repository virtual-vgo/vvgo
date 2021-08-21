package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_valuesToParts(t *testing.T) {
	got := valuesToParts([][]interface{}{
		{"Project", "Score Order", "Part Name", "Sheet Music File", "Click Track File", "Pronunciation Guide", "Conductor Video"},
		{"04-between-heaven-and-earth", 33, "Suspended Cymbal", "32. Between Heaven and Earth - Suspended Cymbal.pdf", "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3", "", "https://www.youtube.com/watch?v=GHnk2BmAFYg"},
		{"04-between-heaven-and-earth", 34, "Harp", "33. Between Heaven and Earth - Harp.pdf", "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3", "", "https://www.youtube.com/watch?v=zBmHNarPvnA"},
		{"04-between-heaven-and-earth", 35, "Piano", "34. Between Heaven and Earth - Piano.pdf", "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3", "", "https://www.youtube.com/watch?v=zBmHNarPvnA"},
	})

	want := Parts{
		{
			Project:        "04-between-heaven-and-earth",
			PartName:       "Suspended Cymbal",
			ScoreOrder:     33,
			SheetMusicFile: "32. Between Heaven and Earth - Suspended Cymbal.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=GHnk2BmAFYg",
			SheetMusicLink: "/download?object=32. Between Heaven and Earth - Suspended Cymbal.pdf",
			ClickTrackLink: "/download?object=VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
		},
		{
			Project:        "04-between-heaven-and-earth",
			PartName:       "Harp",
			ScoreOrder:     34,
			SheetMusicFile: "33. Between Heaven and Earth - Harp.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=zBmHNarPvnA",
			SheetMusicLink: "/download?object=33. Between Heaven and Earth - Harp.pdf",
			ClickTrackLink: "/download?object=VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
		},
		{
			Project:        "04-between-heaven-and-earth",
			PartName:       "Piano",
			ScoreOrder:     35,
			SheetMusicFile: "34. Between Heaven and Earth - Piano.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=zBmHNarPvnA",
			SheetMusicLink: "/download?object=34. Between Heaven and Earth - Piano.pdf",
			ClickTrackLink: "/download?object=VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
		},
	}

	assert.Equal(t, want, got)
}
