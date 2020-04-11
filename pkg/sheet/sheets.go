package sheet

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"time"
)

const DataFile = "__data.json"

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

func (x Sheets) Init() {
	x.PutObject(DataFile, &storage.Object{
		ContentType: "application/json",
		Buffer:      *bytes.NewBuffer([]byte(`[]`)),
	})
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

func (x Sheets) Store(ctx context.Context, sheets []Sheet, pdfBytes []byte) bool {
	// first, validate all the sheets
	if ok := validateSheets(sheets); !ok {
		return false
	}

	// hash the pdf bytes
	fileKey := fmt.Sprintf("%x.pdf", md5.Sum(pdfBytes))
	for i := range sheets {
		sheets[i].FileKey = fileKey
	}

	// store the pdf
	x.PutObject(fileKey, &storage.Object{
		ContentType: "application/pdf",
		Buffer:      *bytes.NewBuffer(pdfBytes),
	})

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

	// store the data file with a revision
	backupName := fmt.Sprintf("%s-%s", DataFile, time.Now().UTC().Format(time.RFC3339))
	if ok := x.PutObject(backupName, &storage.Object{
		ContentType: "application/json",
		Buffer:      buffer,
	}); !ok {
		return false
	}

	// write the data file
	return x.PutObject(DataFile, &storage.Object{
		ContentType: "application/json",
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
