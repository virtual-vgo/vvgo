package sheets

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"strconv"
)

const BucketName = "sheets"

var (
	ErrMissingProject    = fmt.Errorf("missing required field `project`")
	ErrMissingInstrument = fmt.Errorf("missing required field `instrument`")
	ErrMissingPartNumber = fmt.Errorf("missing required field `part_number`")
)

type Sheet struct {
	Project    string `json:"project"`
	Instrument string `json:"instrument"`
	PartNumber int    `json:"part_number"`
}

func NewSheetFromTags(tags storage.Tags) Sheet {
	partNumber, _ := strconv.Atoi(tags["Part-Number"])
	return Sheet{
		Project:    tags["Project"],
		Instrument: tags["Instrument"],
		PartNumber: partNumber,
	}
}

func NewSheetFromRequest(r *http.Request) (Sheet, error) {
	partNumber, _ := strconv.Atoi(r.FormValue("part_number"))
	sheet := Sheet{
		Project:    r.FormValue("project"),
		Instrument: r.FormValue("instrument"),
		PartNumber: partNumber,
	}
	return sheet, sheet.Validate()
}

func (x Sheet) ObjectKey() string {
	return fmt.Sprintf("%s-%s-%d.pdf", x.Project, x.Instrument, x.PartNumber)
}

func (x Sheet) Link() string {
	return fmt.Sprintf("/download?bucket=%s&key=%s", BucketName, x.ObjectKey())
}

func (x Sheet) Tags() map[string]string {
	return map[string]string{
		"Project":     x.Project,
		"Instrument":  x.Instrument,
		"Part-Number": strconv.Itoa(x.PartNumber),
	}
}

func (x Sheet) Validate() error {
	if x.Project == "" {
		return ErrMissingProject
	} else if x.Instrument == "" {
		return ErrMissingInstrument
	} else if x.PartNumber == 0 {
		return ErrMissingPartNumber
	} else {
		return nil
	}
}
