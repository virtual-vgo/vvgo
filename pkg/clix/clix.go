package clix

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
)

const DataFile = "__data.json"

var logger = log.Logger()

var (
	ErrMissingProject    = fmt.Errorf("missing project")
	ErrMissingPartName   = fmt.Errorf("missing part name")
	ErrMissingPartNumber = fmt.Errorf("missing part number")
)

func ValidMediaType(mediaType string) bool {
	return strings.HasPrefix(mediaType, "audio/")
}

type Clix struct {
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

func (x Clix) Init() bool {
	return x.PutObject(DataFile, storage.NewJSONObject(bytes.NewBuffer([]byte(`[]`))))
}

func (x Clix) List() []Click {
	// grab the data file
	var dest storage.Object
	if ok := x.GetObject(DataFile, &dest); !ok {
		return nil
	}

	// deserialize the data file
	var clix []Click
	if err := json.NewDecoder(&dest.Buffer).Decode(&clix); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil
	}
	return clix
}

type File struct {
	MediaType string
	Ext       string
	Bytes     []byte
}

func (x Clix) Store(ctx context.Context, clix []Click, file *File) bool {
	// first, validate all the clix
	if ok := validateClix(clix); !ok {
		return false
	}

	// hash the file bytes
	fileKey := fmt.Sprintf("%x%s", md5.Sum(file.Bytes), file.Ext)
	for i := range clix {
		clix[i].FileKey = fileKey
	}

	// store the file
	x.PutObject(fileKey, storage.NewObject(file.MediaType, bytes.NewBuffer(file.Bytes)))

	// now we update the data file
	// read+modify+write means we need a lock
	if ok := x.Lock(ctx); !ok {
		return false
	}
	defer x.Unlock()

	// pull down the clix data
	allClix := x.List()
	if allClix == nil {
		return false
	}
	allClix = append(allClix, clix...)

	// encode the data file
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&allClix); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}

	return storage.WithBackup(x.PutObject)(DataFile, storage.NewJSONObject(&buffer))
}

func validateClix(clix []Click) bool {
	for _, click := range clix {
		if err := click.Validate(); err != nil {
			logger.WithError(err).Error("click failed validation")
			return false
		}
	}
	return true
}

type Click struct {
	Project    string `json:"project"`
	PartName   string `json:"part_name"`
	PartNumber uint8  `json:"part_number"`
	FileKey    string `json:"file_key"`
}

func (x Click) String() string {
	return fmt.Sprintf("Project: %s Part: %s-%d", x.Project, x.PartName, x.PartNumber)
}

func (x Click) ObjectKey() string {
	return x.FileKey
}

func (x Click) Link(bucket string) string {
	return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.ObjectKey())
}

func (x Click) Validate() error {
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
