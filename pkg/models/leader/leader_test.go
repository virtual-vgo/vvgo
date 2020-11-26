package leader

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValuesToLeaders(t *testing.T) {
	got := valuesToLeaders([][]interface{}{
		{"Name", "Epithet", "Affiliations", "Blurb", "Email", "Icon"},
		{"Brandon", "Keeper of Smol Horn™", "Reno Video Game Symphony, The Intermission Orchestra at Berkeley", "", "brandon@vvgo.org", "images/leaders/brandon-128x128.jpg"},
		{"Jackson", "Coder of Things", "SACWE, TIO @ Berkeley", "I like to make music and code things.", "jackson@vvgo.org", "images/leaders/jackson-128x128.jpg"},
		{"Jacob", "Creator of the Musics", "Zelda Universe", "Making music, having fun, and making more music.", "jacob@vvgo.org", "images/leaders/jacob-128x128.jpg"},
		{"Jerome", "(b.1812 - ", "Awesome Orchestra Collective, Game Music Ensemble @ UCLA, The Giggling Donkey, Golden State Gamer Symphony Orchestra, Hitbox Music Ensemble, Video Games Live", "https://youtu.be/HNK_KB6m6H0", "jerome@vvgo.org", "images/leaders/jerome-128x128.jpg"},
		{"Jose", "", "", "", "", "images/leaders/jose-128x128.jpg"}},
	)

	want := Leaders{
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
	assert.Equal(t, want, got)
}
