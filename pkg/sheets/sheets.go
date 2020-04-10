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

const BucketName = "sheets"
const DataFile = "sheets.json"
const LockName = "sheets.lock"

var logger = log.Logger()

var (
	ErrMissingProject    = fmt.Errorf("missing project")
	ErrMissingPartName   = fmt.Errorf("missing part name")
	ErrMissingPartNumber = fmt.Errorf("missing part number")
)

type Sheet struct {
	Project    string `json:"project"`
	PartName   string `json:"part_name"`
	PartNumber int    `json:"part_number"`
	FileKey    string `json:"file_key"`
}

func (x Sheet) ObjectKey() string {
	return x.FileKey
}

func (x Sheet) Link() string {
	return fmt.Sprintf("/download?bucket=%s&key=%s", BucketName, x.ObjectKey())
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

type Storage struct {
	*storage.RedisLocker
	*storage.MinioDriver
}

func (x *Storage) List() []Sheet {
	// grab the data file
	var dest storage.Object
	if ok := x.GetObject(BucketName, DataFile, &dest); !ok {
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

func (x *Storage) Store(ctx context.Context, sheets []Sheet, pdfBytes []byte) bool {
	// first, validate all the sheets
	if ok := validateSheets(sheets); !ok {
		return false
	}

	// read+modify+write means we need a lock
	lock := x.Lock(ctx, LockName)
	if lock == nil {
		return false
	}
	defer lock.Release()

	// pull down the sheets data
	allSheets := x.List()
	if allSheets == nil {
		return false
	}

	// hash the pdf bytes
	fileKey := fmt.Sprintf("%x.pdf", md5.Sum(pdfBytes))

	// store the pdf
	x.PutObject(BucketName, &storage.Object{
		ContentType: "application/pdf",
		Name:        fileKey,
		Buffer:      *bytes.NewBuffer(pdfBytes),
	})

	// update sheets with the file key
	for i := range sheets {
		sheets[i].FileKey = fileKey
	}
	allSheets = append(allSheets, sheets...)

	// encode the data file
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&allSheets); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}

	// store the data file
	return x.PutObject(BucketName, &storage.Object{
		ContentType: "application/json",
		Name:        DataFile,
		Buffer:      buffer,
	})
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
