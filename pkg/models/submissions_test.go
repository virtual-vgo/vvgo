package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_valuesToSubmissionRecords(t *testing.T) {
	got := valuesToSubmissionRecords([][]interface{}{
		{"Credited Name", "Instrument", "Bottom Text"},
		{"Calem Destiny", "Soprano", "1"},
		{"Cheryl Carr", "Soprano", "1"},
		{"Chris Erickson", "Soprano", "1"},
	})

	want := Submissions{
		{CreditedName: "Calem Destiny", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Cheryl Carr", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Chris Erickson", Instrument: "Soprano", BottomText: "1"},
	}

	assert.Equal(t, want, got)
}

func TestSubmissions_ToCredits(t *testing.T) {
	got := Submissions{
		{CreditedName: "Calem Destiny", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Cheryl Carr", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Chris Erickson", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Karenna Foley", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Lena Świt", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Lucia C.", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Mashica Washington", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "May Claire La Plante", Instrument: "Soprano", BottomText: "1"},
		{CreditedName: "Caroline Augelli", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Cheryl Carr", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Chris Erickson", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Gabrielle B.", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Ian Martyn", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Joie Zhou", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Lena Świt", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Lucia C.", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Mashica Washington", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "May Claire La Plante", Instrument: "Soprano", BottomText: "2"},
		{CreditedName: "Brandon Harnish", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Caroline Augelli", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Cheryl Carr", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Gabrielle B.", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Ian Martyn", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Joie Zhou", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Lena Świt", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Lucia C.", Instrument: "Alto", BottomText: "1"},
		{CreditedName: "Azure", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Brandon Harnish", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Caroline Augelli", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Frances Lee", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Gabrielle B.", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Ian Martyn", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Jessica Muñoz", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Joie Zhou", Instrument: "Alto", BottomText: "2"},
		{CreditedName: "Brandon Harnish", Instrument: "Tenor", BottomText: "1"},
		{CreditedName: "Calem Destiny", Instrument: "Tenor", BottomText: "1"},
	}.ToCredits("06-aurene-dragon-full-of-light")

	want := Credits{
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
	}

	assert.Equal(t, want, got)
}
