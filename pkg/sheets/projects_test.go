package sheets

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_valuesToProjects(t *testing.T) {
	got := valuesToProjects([][]interface{}{
		{"Name", "Title", "Parts Released", "Parts Archived", "Season", "Submission Link", "Submission Deadline", "Sources", "Composers", "Arrangers", "Editors/Adaptors", "Transcribers", "Preparers", "Clix By", "Reviewers", "Lyricists", "Additional Content", "Reference Track", "Choir Pronunciation Guide", "Youtube Link", "Banner Link", "Youtube Embed", "Bandcamp Embed"},
		{"01-snake-eater", "Snake Eater", true, true, "", "", "", "Metal Gear Solid 3", "Norihiko Hibino (日比野 則彦)", "Edited by Jerome Landingin", "Jerome Landingin", "", "The Giggling Donkey, Inc.", "Finny Jacob Zeleny", "", "", "Brandon Harnish", "01_MSG3_Snake-Eater_VVGO_Reference.mp3", "", "https://bit.ly/vvgo01", "/images/snake-eater-title-text.png", "https://www.youtube.com/embed/HVKRro_lizk", "https://bandcamp.com/EmbeddedPlayer/album=3483061952/size=large/bgcol=333333/linkcol=9a64ff/tracklist=false/artwork=small/transparent=true/"},
		{"02-proof-of-a-hero", "Proof of a Hero", true, true, "", "", "", "Monster Hunter", "Masato Kouda (甲田 雅人)", "Arranged by Jacob Zeleny", "", "Jacob Zeleny", "The Giggling Donkey, Inc., Thomas Håkanson", "Jacob Zeleny", "Brandon Harnish", "", "", "02_MH_Proof-of-a-Hero_Reference-Track_W-CLIX", "", "https://bit.ly/vvgo2", "/images/Site_Banner_-_Proof_of_a_Hero.png", "https://www.youtube.com/embed/GJZtTe7Ayks", ""},
		{"03-the-end-begins-to-rock", "The End Begins (To Rock)", true, true, "", "", "", "God of War II & Guitar Hero III", "Gerard K. Marino", "Orch. Shota Nakama; Additional Orch. & Arr. Thomas Håkanson", "", "", "Thomas Håkanson", "Jacob Zeleny", "Brandon Harnish, Elliot McAuley, Jerome Landingin, Thomas Håkanson", "", "", "03_The-End-Begins-to-Rock_Reference-Track-NoCLIX.mp3", "", "https://bit.ly/vvgo03", "/images/VVGO_03_TEBTR_Website_Title.png", "https://www.youtube.com/embed/2V52as93SEE", ""},
	})

	want := Projects{
		{
			Name:              "01-snake-eater",
			Title:             "Snake Eater",
			PartsReleased:     true,
			PartsArchived:     true,
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
			PartsReleased:  true,
			PartsArchived:  true,
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
			PartsReleased:  true,
			PartsArchived:  true,
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
	assert.Equal(t, want, got)
}

func TestProjects_Query(t *testing.T) {
	assert.Equal(t, Projects{
		{
			Name:          "01-snake-eater",
			Title:         "Snake Eater",
			PartsReleased: true,
			PartsArchived: true,
			YoutubeLink:   "https://bit.ly/vvgo01",
			YoutubeEmbed:  "https://www.youtube.com/embed/HVKRro_lizk",
			BannerLink:    "/images/snake-eater-title-text.png",
		},
	}, Projects{
		{
			Name:          "01-snake-eater",
			Title:         "Snake Eater",
			PartsReleased: true,
			PartsArchived: true,
			YoutubeLink:   "https://bit.ly/vvgo01",
			YoutubeEmbed:  "https://www.youtube.com/embed/HVKRro_lizk",
			BannerLink:    "/images/snake-eater-title-text.png",
		},
		{
			Name:          "02-proof-of-a-hero",
			Title:         "Proof of a Hero",
			PartsReleased: true,
			PartsArchived: false,
			YoutubeLink:   "https://bit.ly/vvgo2",
			YoutubeEmbed:  "https://www.youtube.com/embed/GJZtTe7Ayks",
			BannerLink:    "/images/Site_Banner_-_Proof_of_a_Hero.png",
		},
	}.Query(map[string]interface{}{"Parts Released": true, "Parts Archived": true}))
}
