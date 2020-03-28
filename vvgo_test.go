package main

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	"net/http"
	"net/url"
	"testing"
)

func TestApiServer_Index(t *testing.T) {
	type fields struct {
		MinioClient *minio.Client
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	for _, tt := range []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	} {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestApiServer_Upload(t *testing.T) {
	type fields struct {
		MinioClient *minio.Client
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	for _, tt := range []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	} {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestMusicPDFMeta_ReadFromUrlValues(t *testing.T) {
	values, err := url.ParseQuery(`project=test-project&instrument=test-instrument&part_number=4`)
	if err != nil {
		t.Fatalf("url.ParseQuery() failed: %v", err)
	}

	expectedMeta := MusicPDFMeta{
		Project:    "test-project",
		Instrument: "test-instrument",
		PartNumber: 4,
	}

	var gotMeta MusicPDFMeta
	gotMeta.ReadFromUrlValues(values)

	if expected, got := fmt.Sprintf("%#v", expectedMeta), fmt.Sprintf("%#v", gotMeta); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestMusicPDFMeta_Validate(t *testing.T) {
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
			x := &MusicPDFMeta{
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
