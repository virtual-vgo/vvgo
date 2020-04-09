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

var logger = log.Logger()

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
