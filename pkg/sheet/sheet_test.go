package sheet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSheet_Link(t *testing.T) {
	sheet := Sheet{FileKey: "mock-file-key"}
	assert.Equal(t, sheet.Link("sheets"), "/download?bucket=sheets&object=mock-file-key")
}

func TestSheet_ObjectKey(t *testing.T) {
	sheet := Sheet{FileKey: "mock-file-key"}
	assert.Equal(t, sheet.ObjectKey(), "mock-file-key")
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
			want: ErrMissingPartName,
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
				PartName:   tt.fields.Instrument,
				PartNumber: tt.fields.PartNumber,
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}
