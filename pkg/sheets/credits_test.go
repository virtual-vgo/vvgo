package sheets

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCredits_WebsitePasta(t *testing.T) {
	got := Credits{
		{Project: "06-aurene-dragon-full-of-light", Order: 0, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Calem Destiny", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 1, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Cheryl Carr", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 2, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Chris Erickson", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 3, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Karenna Foley", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 4, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lena Świt", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 5, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lucia C.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 6, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Mashica Washington", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 7, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "May Claire La Plante", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 8, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Caroline Augelli", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 11, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Gabrielle B.", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 12, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Ian Martyn", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 13, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Joie Zhou", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 18, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Brandon Harnish", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 19, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Caroline Augelli", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 20, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Cheryl Carr", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 21, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Gabrielle B.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 22, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Ian Martyn", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 23, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Joie Zhou", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 24, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lena Świt", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 25, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lucia C.", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 26, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Azure", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 29, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Frances Lee", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 32, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Jessica Muñoz", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 34, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Brandon Harnish", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 35, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Calem Destiny", BottomText: "(1)"},
	}.WebsitePasta()

	want := `06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Calem Destiny	(1)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Cheryl Carr	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Chris Erickson	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Karenna Foley	(1)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Lena Świt	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Lucia C.	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Mashica Washington	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	May Claire La Plante	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Caroline Augelli	(2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Gabrielle B.	(2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Ian Martyn	(2)
06-aurene-dragon-full-of-light		PERFORMERS	SOPRANO	Joie Zhou	(2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Brandon Harnish	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Caroline Augelli	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Cheryl Carr	(1)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Gabrielle B.	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Ian Martyn	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Joie Zhou	(1, 2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Lena Świt	(1)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Lucia C.	(1)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Azure	(2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Frances Lee	(2)
06-aurene-dragon-full-of-light		PERFORMERS	ALTO	Jessica Muñoz	(2)
06-aurene-dragon-full-of-light		PERFORMERS	TENOR	Brandon Harnish	(1)
06-aurene-dragon-full-of-light		PERFORMERS	TENOR	Calem Destiny	(1)
`

	assert.Equal(t, want, got)
}

func TestCredits_VideoPasta(t *testing.T) {
	got := Credits{
		{Project: "06-aurene-dragon-full-of-light", Order: 0, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Calem Destiny", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 1, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Cheryl Carr", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 2, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Chris Erickson", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 3, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Karenna Foley", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 4, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lena Świt", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 5, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lucia C.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 6, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Mashica Washington", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 7, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "May Claire La Plante", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 8, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Caroline Augelli", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 11, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Gabrielle B.", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 12, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Ian Martyn", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 13, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Joie Zhou", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 18, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Brandon Harnish", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 19, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Caroline Augelli", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 20, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Cheryl Carr", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 21, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Gabrielle B.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 22, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Ian Martyn", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 23, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Joie Zhou", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 24, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lena Świt", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 25, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lucia C.", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 26, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Azure", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 29, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Frances Lee", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 32, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Jessica Muñoz", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 34, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Brandon Harnish", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 35, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Calem Destiny", BottomText: "(1)"},
	}.VideoPasta()

	want := `— PERFORMERS —	— PERFORMERS —
SOPRANO	SOPRANO	Calem Destiny (1)	Cheryl Carr (1, 2)	Chris Erickson (1, 2)	Karenna Foley (1)	Lena Świt (1, 2)	Lucia C. (1, 2)	Mashica Washington (1, 2)	May Claire La Plante (1, 2)	Caroline Augelli (2)	Gabrielle B. (2)	Ian Martyn (2)	Joie Zhou (2)
ALTO	ALTO	Brandon Harnish (1, 2)	Caroline Augelli (1, 2)	Cheryl Carr (1)	Gabrielle B. (1, 2)	Ian Martyn (1, 2)	Joie Zhou (1, 2)	Lena Świt (1)	Lucia C. (1)	Azure (2)	Frances Lee (2)	Jessica Muñoz (2)
TENOR	TENOR	Brandon Harnish (1)	Calem Destiny (1)
`
	assert.Equal(t, want, got)
}

func TestCredits_YoutubePasta(t *testing.T) {
	got := Credits{
		{Project: "06-aurene-dragon-full-of-light", Order: 0, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Calem Destiny", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 1, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Cheryl Carr", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 2, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Chris Erickson", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 3, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Karenna Foley", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 4, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lena Świt", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 5, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Lucia C.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 6, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Mashica Washington", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 7, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "May Claire La Plante", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 8, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Caroline Augelli", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 11, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Gabrielle B.", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 12, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Ian Martyn", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 13, MajorCategory: "PERFORMERS", MinorCategory: "SOPRANO", Name: "Joie Zhou", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 18, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Brandon Harnish", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 19, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Caroline Augelli", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 20, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Cheryl Carr", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 21, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Gabrielle B.", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 22, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Ian Martyn", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 23, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Joie Zhou", BottomText: "(1, 2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 24, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lena Świt", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 25, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Lucia C.", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 26, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Azure", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 29, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Frances Lee", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 32, MajorCategory: "PERFORMERS", MinorCategory: "ALTO", Name: "Jessica Muñoz", BottomText: "(2)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 34, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Brandon Harnish", BottomText: "(1)"},
		{Project: "06-aurene-dragon-full-of-light", Order: 35, MajorCategory: "PERFORMERS", MinorCategory: "TENOR", Name: "Calem Destiny", BottomText: "(1)"},
	}.YoutubePasta()

	want := `— PERFORMERS —

SOPRANO
Calem Destiny (1), Cheryl Carr (1, 2), Chris Erickson (1, 2), Karenna Foley (1), Lena Świt (1, 2), Lucia C. (1, 2), Mashica Washington (1, 2), May Claire La Plante (1, 2), Caroline Augelli (2), Gabrielle B. (2), Ian Martyn (2), Joie Zhou (2)

ALTO
Brandon Harnish (1, 2), Caroline Augelli (1, 2), Cheryl Carr (1), Gabrielle B. (1, 2), Ian Martyn (1, 2), Joie Zhou (1, 2), Lena Świt (1), Lucia C. (1), Azure (2), Frances Lee (2), Jessica Muñoz (2)

TENOR
Brandon Harnish (1), Calem Destiny (1)
`
	assert.Equal(t, want, got)
}

func Test_valuesToCredits(t *testing.T) {
	got := valuesToCredits([][]interface{}{
		{"Project", "Order", "Major Category", "Minor Category", "Name", "Bottom Text"},
		{"01-snake-eater", 16, "CREW", "SCORE PREPARATION", "The Giggling Donkey,", "INC."},
		{"01-snake-eater", 17, "CREW", "SCORE PREPARATION", "Brandon Harnish", "(CHORAL SCORE)"},
		{"01-snake-eater", 18, "CREW", "ENGRAVING", "The Giggling Donkey,", "INC."},
	})

	want := []Credit{
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

	assert.Equal(t, want, got)
}
