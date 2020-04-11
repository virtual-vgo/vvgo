package sheet

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
)

var logger = log.Logger()

var (
	ErrMissingProject    = fmt.Errorf("missing project")
	ErrMissingPartName   = fmt.Errorf("missing part name")
	ErrMissingPartNumber = fmt.Errorf("missing part number")
)

type Sheet struct {
	Project    string `json:"project"`
	PartName   string `json:"part_name"`
	PartNumber uint   `json:"part_number"`
	FileKey    string `json:"file_key"`
}

func (x Sheet) String() string {
	return fmt.Sprintf("Project: %s Part: %s-%d", x.Project, x.PartName, x.PartNumber)
}

func (x Sheet) ObjectKey() string {
	return x.FileKey
}

func (x Sheet) Link(bucket string) string {
	return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.ObjectKey())
}

func (x Sheet) Validate() error {
	if x.Project == "" {
		return ErrMissingProject
	} else if x.PartName == "" {
		return ErrMissingPartName
	} else if x.PartNumber == 0 {
		return ErrMissingPartNumber
	} else {
		return nil
	}
}
