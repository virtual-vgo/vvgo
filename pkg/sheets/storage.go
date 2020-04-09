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
	Locker
	ObjectStorage
}

type ObjectStorage interface {
	GetObject(bucketName string, key string, object *storage.Object) bool
	PutObject(bucketName string, object *storage.Object) bool
}

type Locker interface {
	Lock(ctx context.Context, name string) Lock
}

type Lock interface {
	Release() error
}

func (x *Storage) Store(ctx context.Context, sheets []Sheet, pdfBytes []byte) bool {
	// read+modify+write means we need a lock
	lock := x.Lock(ctx, LockName)
	if lock == nil {
		return false
	}
	defer lock.Release()

	// grab the data file
	var dest storage.Object
	if ok := x.ObjectStorage.GetObject(BucketName, DataFile, &dest); !ok {
		return false
	}

	// deserialize the data file
	var allSheets []Sheet
	if err := json.NewDecoder(&dest.Buffer).Decode(&allSheets); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
	}

	// hash the pdf bytes
	fileKey := fmt.Sprintf("%x.pdf", md5.Sum(pdfBytes))

	// store the pdf
	x.ObjectStorage.PutObject(BucketName, &storage.Object{
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

	// write the data file
	return x.ObjectStorage.PutObject(BucketName, &storage.Object{
		ContentType: "application/json",
		Name:        DataFile,
		Buffer:      buffer,
	})
}
