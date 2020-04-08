package sheets

import (
	"fmt"
	"testing"
)

func TestNewSheetFromTags(t *testing.T) {
	tags := map[string]string{
		"Project":     "01-snake-eater",
		"Instrument":  "trumpet",
		"Part-Number": "4",
	}

	expectedMeta := Sheet{
		Project:    "01-snake-eater",
		Instrument: "trumpet",
		PartNumber: 4,
	}

	gotMeta := NewSheetFromTags(tags)
	if expected, got := fmt.Sprintf("%#v", expectedMeta), fmt.Sprintf("%#v", gotMeta); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestSheet_ToTags(t *testing.T) {
	meta := Sheet{
		Project:    "01-snake-eater",
		Instrument: "trumpet",
		PartNumber: 4,
	}

	wantMap := map[string]string{
		"Project":     "01-snake-eater",
		"Instrument":  "trumpet",
		"Part-Number": "4",
	}
	gotMap := meta.Tags()
	if expected, got := fmt.Sprintf("%#v", wantMap), fmt.Sprintf("%#v", gotMap); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestSheet_Validate(t *testing.T) {
	type fields struct {
		Project    string
		Instrument string
		PartNumber int
	}
	for _, tt := range []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "valid",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: nil,
		},
		{
			name: "missing project",
			fields: fields{
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: ErrMissingProject,
		},
		{
			name: "missing instrument",
			fields: fields{
				Project:    "test-project",
				PartNumber: 6,
			},
			want: ErrMissingInstrument,
		},
		{
			name: "missing part number",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
			},
			want: ErrMissingPartNumber,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := &Sheet{
				Project:    tt.fields.Project,
				Instrument: tt.fields.Instrument,
				PartNumber: tt.fields.PartNumber,
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}
