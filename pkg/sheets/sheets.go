package sheets

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
)

const DataFile = "__data.json"

var logger = log.Logger()

var (
	ErrMissingProject    = fmt.Errorf("missing project")
	ErrMissingPartName   = fmt.Errorf("missing part name")
	ErrMissingPartNumber = fmt.Errorf("missing part number")
)

func ValidMediaType(mediaType string) bool {
	return mediaType == "application/pdf"
}

type Sheets struct {
	Bucket
	Locker
}

type Bucket interface {
	PutObject(name string, object *storage.Object) bool
	GetObject(name string, dest *storage.Object) bool
	DownloadURL(name string) (string, error)
}

type Locker interface {
	Lock(ctx context.Context) bool
	Unlock()
}

func (x Sheets) Init() bool {
	return x.PutObject(DataFile, storage.NewJSONObject(bytes.NewBuffer([]byte(`[]`))))
}

func (x Sheets) List() []Sheet {
	// grab the data file
	var dest storage.Object
	if ok := x.GetObject(DataFile, &dest); !ok {
		return nil
	}

	// deserialize the data file
	var sheets []Sheet
	if err := json.NewDecoder(&dest.Buffer).Decode(&sheets); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil
	}
	return sheets
}

func (x Sheets) Store(ctx context.Context, sheets []Sheet, fileBytes []byte) bool {
	// first, validate all the sheets
	if ok := validateSheets(sheets); !ok {
		return false
	}

	// hash the pdf bytes
	fileKey := fmt.Sprintf("%x.pdf", md5.Sum(fileBytes))
	for i := range sheets {
		sheets[i].FileKey = fileKey
	}

	// store the pdf
	x.PutObject(fileKey, storage.NewObject("application/pdf", bytes.NewBuffer(fileBytes)))

	// now we update the data file
	// read+modify+write means we need a lock
	if ok := x.Lock(ctx); !ok {
		return false
	}
	defer x.Unlock()

	// pull down the sheets data
	allSheets := x.List()
	if allSheets == nil {
		return false
	}
	allSheets = append(allSheets, sheets...)

	// encode the data file
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&allSheets); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}

	return storage.WithBackup(x.PutObject)(DataFile, storage.NewJSONObject(&buffer))
}

func validateSheets(sheets []Sheet) bool {
	for _, sheet := range sheets {
		if err := sheet.Validate(); err != nil {
			logger.WithError(err).Error("sheet failed validation")
			return false
		}
	}
	return true
}

type Sheet struct {
	Project    string `json:"project"`
	PartName   string `json:"part_name"`
	PartNumber uint8  `json:"part_number"`
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
