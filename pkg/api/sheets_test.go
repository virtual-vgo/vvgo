package api

import (
	"encoding/csv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
)

func TestValuesToParts(t *testing.T) {
	expected := []Part{
		{
			Project:        "04-between-heaven-and-earth",
			ProjectTitle:   "Between Heaven and Earth",
			PartName:       "Suspended Cymbal",
			ScoreOrder:     33,
			SheetMusicFile: "32. Between Heaven and Earth - Suspended Cymbal.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=GHnk2BmAFYg",
			Released:       true,
			Archived:       true,
			ReferenceTrack: "VVGO 04 FE3H Between Heaven and Earth - Reference Track.mp3",
		},
		{
			Project:        "04-between-heaven-and-earth",
			ProjectTitle:   "Between Heaven and Earth",
			PartName:       "Harp",
			ScoreOrder:     34,
			SheetMusicFile: "33. Between Heaven and Earth - Harp.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=zBmHNarPvnA",
			Released:       true,
			Archived:       true,
			ReferenceTrack: "VVGO 04 FE3H Between Heaven and Earth - Reference Track.mp3",
		},
		{
			Project:        "04-between-heaven-and-earth",
			ProjectTitle:   "Between Heaven and Earth",
			PartName:       "Piano",
			ScoreOrder:     35,
			SheetMusicFile: "34. Between Heaven and Earth - Piano.pdf",
			ClickTrackFile: "VVGO 04 FE3H Between Heaven and Earth - CLIX Track.mp3",
			ConductorVideo: "https://www.youtube.com/watch?v=zBmHNarPvnA",
			Released:       true,
			Archived:       true,
			ReferenceTrack: "VVGO 04 FE3H Between Heaven and Earth - Reference Track.mp3",
		},
	}

	got := ValuesToParts(readValuesFromFile(t, "testdata/parts.tsv", '\t'))
	assert.Equal(t, expected, got)
}

func TestValuesToProjects(t *testing.T) {
	expected := []Project{
		{
			Name:              "01-snake-eater",
			Title:             "Snake Eater",
			Released:          true,
			Archived:          true,
			Sources:           "Metal Gear Solid 3",
			Composers:         "Norihiko Hibino (日比野 則彦)",
			Arrangers:         "Edited by Jerome Landingin",
			Preparers:         "The Giggling Donkey, Inc.",
			ClixBy:            "Finny Jacob Zeleny",
			AdditionalContent: "Brandon Harnish",
			ReferenceTrack:    "01_MSG3_Snake-Eater_VVGO_Reference.mp3",
			YoutubeLink:       "https://bit.ly/vvgo01",
			YoutubeEmbed:      "https://www.youtube.com/embed/HVKRro_lizk",
			BannerLink:        "/images/snake-eater-title-text.png",
		},
		{
			Name:           "02-proof-of-a-hero",
			Title:          "Proof of a Hero",
			Released:       true,
			Archived:       true,
			Sources:        "Monster Hunter",
			Composers:      "Masato Kouda (甲田 雅人)",
			Arrangers:      "Arranged by Jacob Zeleny",
			Transcribers:   "Jacob Zeleny",
			Preparers:      "The Giggling Donkey, Inc., Thomas Håkanson",
			ClixBy:         "Jacob Zeleny",
			Reviewers:      "Brandon Harnish",
			ReferenceTrack: "02_MH_Proof-of-a-Hero_Reference-Track_W-CLIX",
			YoutubeLink:    "https://bit.ly/vvgo2",
			YoutubeEmbed:   "https://www.youtube.com/embed/GJZtTe7Ayks",
			BannerLink:     "/images/Site_Banner_-_Proof_of_a_Hero.png",
		},
		{
			Name:           "03-the-end-begins-to-rock",
			Title:          "The End Begins (To Rock)",
			Released:       true,
			Archived:       true,
			Sources:        "God of War II & Guitar Hero III",
			Composers:      "Gerard K. Marino",
			Arrangers:      "Orch. Shota Nakama; Additional Orch. & Arr. Thomas Håkanson",
			Preparers:      "Thomas Håkanson",
			ClixBy:         "Jacob Zeleny",
			Reviewers:      "Brandon Harnish, Elliot McAuley, Jerome Landingin, Thomas Håkanson",
			ReferenceTrack: "03_The-End-Begins-to-Rock_Reference-Track-NoCLIX.mp3",
			YoutubeLink:    "https://bit.ly/vvgo03",
			YoutubeEmbed:   "https://www.youtube.com/embed/2V52as93SEE",
			BannerLink:     "/images/VVGO_03_TEBTR_Website_Title.png",
		},
	}

	got := ValuesToProjects(readValuesFromFile(t, "testdata/projects.tsv", '\t'))
	assert.Equal(t, expected, got)
}

func TestValuesToCredits(t *testing.T) {
	expected := []Credit{
		{
			Project:       "01-snake-eater",
			Order:         16,
			MajorCategory: "CREW",
			MinorCategory: "SCORE PREPARATION",
			Name:          "The Giggling Donkey,",
			BottomText:    "INC.",
		},
		{
			Project:       "01-snake-eater",
			Order:         17,
			MajorCategory: "CREW",
			MinorCategory: "SCORE PREPARATION",
			Name:          "Brandon Harnish",
			BottomText:    "(CHORAL SCORE)",
		},
		{
			Project:       "01-snake-eater",
			Order:         18,
			MajorCategory: "CREW",
			MinorCategory: "ENGRAVING",
			Name:          "The Giggling Donkey,",
			BottomText:    "INC.",
		},
	}

	got := ValuesToCredits(readValuesFromFile(t, "testdata/credits.tsv", '\t'))
	assert.Equal(t, expected, got)
}

func TestValuesToLeaders(t *testing.T) {
	expected := []Leader{
		{
			Name:         "Brandon",
			Epithet:      "Keeper of Smol Horn™",
			Affiliations: "Reno Video Game Symphony, The Intermission Orchestra at Berkeley",
			Icon:         "images/leaders/brandon-128x128.jpg",
			Email:        "brandon@vvgo.org",
		},
		{
			Name:         "Jackson",
			Epithet:      "Coder of Things",
			Affiliations: "SACWE, TIO @ Berkeley",
			Blurb:        "I like to make music and code things.",
			Icon:         "images/leaders/jackson-128x128.jpg",
			Email:        "jackson@vvgo.org",
		},
		{
			Name:         "Jacob",
			Epithet:      "Creator of the Musics",
			Affiliations: "Zelda Universe",
			Blurb:        "Making music, having fun, and making more music.",
			Icon:         "images/leaders/jacob-128x128.jpg",
			Email:        "jacob@vvgo.org",
		},
		{
			Name:         "Jerome",
			Epithet:      "(b.1812 - ",
			Affiliations: "Awesome Orchestra Collective, Game Music Ensemble @ UCLA, The Giggling Donkey, Golden State Gamer Symphony Orchestra, Hitbox Music Ensemble, Video Games Live",
			Blurb:        "https://youtu.be/HNK_KB6m6H0",
			Icon:         "images/leaders/jerome-128x128.jpg",
			Email:        "jerome@vvgo.org",
		},
		{
			Name: "Jose",
			Icon: "images/leaders/jose-128x128.jpg",
		},
	}

	got := ValuesToLeaders(readValuesFromFile(t, "testdata/leaders.tsv", '\t'))
	assert.Equal(t, expected, got)
}

func readValuesFromFile(t *testing.T, path string, separator rune) [][]interface{} {
	in, err := os.Open(path)
	require.NoError(t, err, "os.Open() failed")
	defer in.Close()
	reader := csv.NewReader(in)
	reader.Comma = separator
	csvValues, err := reader.ReadAll()
	require.NoError(t, err, "csvReader.ReadAll() failed")
	values := make([][]interface{}, len(csvValues))
	for i, row := range csvValues {
		values[i] = make([]interface{}, len(row))
		for j, col := range row {
			intVal, err := strconv.ParseInt(col, 10, 64)
			if err == nil {
				values[i][j] = intVal
				continue
			}
			boolVal, err := strconv.ParseBool(col)
			if err == nil {
				values[i][j] = boolVal
				continue
			}
			values[i][j] = col
		}
	}
	return values
}
