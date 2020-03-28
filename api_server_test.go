package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestApiServer_Index(t *testing.T) {
	type wants struct {
		code int
		body string
	}

	for _, tt := range []struct {
		name        string
		objectStore MockObjectStore
		request     *http.Request
		wants       wants
	}{
		{
			name:    "method post",
			request: httptest.NewRequest(http.MethodPost, "/", nil),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name: "method get",
			objectStore: MockObjectStore{
				listObjects: func(bucketName string) []Object {
					return []Object{
						{
							ContentType: "application/pdf",
							Name:        "trumpet.pdf",
							Meta: http.Header{
								"instrument": []string{"trumpet"},
							},
						},
						{
							ContentType: "application/pdf",
							Name:        "flute.pdf",
							Meta: http.Header{
								"instrument": []string{"flute"},
							},
						},
					}
				},
			},
			request: httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			wants: wants{
				code: http.StatusOK,
				body: `[{"content-type":"application/pdf","name":"trumpet.pdf","meta":{"instrument":["trumpet"]}},{"content-type":"application/pdf","name":"flute.pdf","meta":{"instrument":["flute"]}}]`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse := httptest.NewRecorder()
			apiServer := ApiServer{tt.objectStore}
			apiServer.Index(gotResponse, tt.request)
			if expected, got := tt.wants.code, gotResponse.Code; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, strings.TrimSpace(gotResponse.Body.String()); expected != got {
				t.Errorf("expected body `%s`, got `%s`", expected, got)
			}
		})
	}
}

func TestApiServer_Upload(t *testing.T) {
	type wants struct {
		code   int
		object Object
	}
	for _, tt := range []struct {
		name        string
		objectStore MockObjectStore
		request     *http.Request
		wants       wants
	}{
		{
			name:    "get",
			request: httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name:    "post with too large body",
			request: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, 1e6+1))),
			wants:   wants{code: http.StatusRequestEntityTooLarge},
		},
		{
			name:    "post with missing fields",
			request: httptest.NewRequest(http.MethodPost, "/?project=test-project&instrument=test-instrument", strings.NewReader("")),
			wants:   wants{code: http.StatusBadRequest},
		},
		{
			name: "post with db error",
			objectStore: MockObjectStore{
				putObject: func(string, *Object) error {
					return fmt.Errorf("mock error")
				},
			},
			request: httptest.NewRequest(http.MethodPost, "/?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader(":wave:")),
			wants: wants{
				object: Object{
					ContentType: "application/pdf",
					Name:        "01-snake-eater-trumpet-4.pdf",
					Meta: http.Header{
						"Project":    []string{"01-snake-eater"},
						"Instrument": []string{"trumpet"},
						"Partnumber": []string{"4"},
					},
					Buffer: *bytes.NewBufferString(":wave:"),
				},
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "success",
			objectStore: MockObjectStore{
				putObject: func(string, *Object) error {
					return nil
				},
			},
			request: httptest.NewRequest(http.MethodPost, "/?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader(":wave:")),
			wants: wants{
				object: Object{
					ContentType: "application/pdf",
					Name:        "01-snake-eater-trumpet-4.pdf",
					Meta: http.Header{
						"Project":    []string{"01-snake-eater"},
						"Instrument": []string{"trumpet"},
						"Partnumber": []string{"4"},
					},
					Buffer: *bytes.NewBufferString(":wave:"),
				},
				code: http.StatusOK,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotObject Object
			gotResponse := httptest.NewRecorder()
			apiServer := ApiServer{MockObjectStore{
				listObjects: tt.objectStore.listObjects,
				putObject: func(bucketName string, object *Object) error {
					gotObject = *object
					return tt.objectStore.putObject(bucketName, object)
				},
			}}
			apiServer.Upload(gotResponse, tt.request)
			if expected, got := tt.wants.code, gotResponse.Code; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
			if expected, got := fmt.Sprintf("%#v", tt.wants.object), fmt.Sprintf("%#v", gotObject); expected != got {
				t.Errorf("\nwant body:%v\n got body:%v", expected, got)
			}
		})
	}
}

func TestMusicPDFMeta_ToHeader(t *testing.T) {
	meta := MusicPDFMeta{
		Project:    "test-project",
		Instrument: "test-instrument",
		PartNumber: 4,
	}

	wantHeader := make(http.Header)
	wantHeader.Add("Project", "test-project")
	wantHeader.Add("Instrument", "test-instrument")
	wantHeader.Add("PartNumber", "4")

	gotHeader := meta.ToHeader()
	if expected, got := fmt.Sprintf("%#v", wantHeader), fmt.Sprintf("%#v", gotHeader); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestMusicPDFMeta_ReadFromHeader(t *testing.T) {
	header := make(http.Header)
	header.Add("Project", "test-project")
	header.Add("Instrument", "test-instrument")
	header.Add("PartNumber", "4")

	expectedMeta := MusicPDFMeta{
		Project:    "test-project",
		Instrument: "test-instrument",
		PartNumber: 4,
	}

	var gotMeta MusicPDFMeta
	gotMeta.ReadFromHeader(header)
	if expected, got := fmt.Sprintf("%#v", expectedMeta), fmt.Sprintf("%#v", gotMeta); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
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
