package parts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/data"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"time"
)

const DataFile = "parts.json"

var logger = log.Logger()

var (
	ErrInvalidPartName   = fmt.Errorf("invalid part name")
	ErrInvalidPartNumber = fmt.Errorf("invalid part number")
)

type Bucket interface {
	PutObject(name string, object *storage.Object) bool
	GetObject(name string, dest *storage.Object) bool
}

type Locker interface {
	Lock(ctx context.Context) bool
	Unlock()
}

type Parts struct {
	Bucket
	Locker
}

func (x Parts) Init() bool {
	return x.PutObject(DataFile, storage.NewJSONObject(bytes.NewBuffer([]byte(`[]`))))
}

func (x Parts) List() []Part {
	// grab the data file
	var dest storage.Object
	if ok := x.GetObject(DataFile, &dest); !ok {
		return nil
	}

	// deserialize the data file
	var parts []Part
	if err := json.NewDecoder(&dest.Buffer).Decode(&parts); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return nil
	}
	return parts
}

func (x Parts) Save(ctx context.Context, parts []Part) bool {
	// first, validate all the parts
	if ok := validatePart(parts); !ok {
		return false
	}

	// now we update the data file
	// read+modify+write means we need a lock
	if ok := x.Lock(ctx); !ok {
		return false
	}
	defer x.Unlock()

	// pull down the parts data
	allParts := x.List()
	if allParts == nil {
		return false
	}

	// merge the changes
	allParts = mergeChanges(allParts, parts)

	// encode the data file
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&allParts); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}

	return storage.WithBackup(x.PutObject)(DataFile, storage.NewJSONObject(&buffer))
}

func mergeChanges(src []Part, changes []Part) []Part {
	// merge the changes
	for _, change := range changes {
		var ok bool
		for i := range src {
			if change.ID.String() == src[i].ID.String() {
				ok = true
				// prepend the new fields
				src[i].Sheets = append(change.Sheets, src[i].Sheets...)
				src[i].Clix = append(change.Clix, src[i].Clix...)
				break
			}
		}
		if !ok {
			src = append(src, change)
		}
	}
	return src
}

func validatePart(parts []Part) bool {
	for _, part := range parts {
		if err := part.Validate(); err != nil {
			logger.WithError(err).Error("part failed validation")
			return false
		}
	}
	return true
}

type Part struct {
	ID
	Clix   Links `json:"click,omitempty"`
	Sheets Links `json:"sheet,omitempty"`
}

type Link struct {
	ObjectKey string
	CreatedAt time.Time
}

type Links []Link

func (x *Links) Key() string { return (*x)[0].ObjectKey }
func (x *Links) NewKey(key string) {
	*x = append([]Link{{
		ObjectKey: key,
		CreatedAt: time.Now(),
	}}, *x...)
}

func (x Part) String() string {
	return fmt.Sprintf("Project: %s Part: %s #%d", x.Project, strings.Title(x.Name), x.Number)
}

func (x Part) SheetLink(bucket string) string {
	if bucket == "" || len(x.Sheets) == 0 {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.Sheets[0].ObjectKey)
	}
}

func (x Part) ClickLink(bucket string) string {
	if bucket == "" || len(x.Clix) == 0 {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.Clix.Key())
	}
}

func (x Part) Validate() error {
	switch true {
	case projects.Exists(x.Project) == false:
		return projects.ErrNotFound
	case ValidNames(x.Name) == false:
		return ErrInvalidPartName
	case ValidNumbers(x.Number) == false:
		return ErrInvalidPartNumber
	default:
		return nil
	}
}

func ValidNames(names ...string) bool {
	for _, name := range names {
		if name == "" {
			return false
		}
		if _, ok := data.ValidPartNames()[name]; !ok {
			return false
		}
	}
	return true
}

func ValidNumbers(numbers ...uint8) bool {
	for _, n := range numbers {
		if n == 0 {
			return false
		}
	}
	return true
}

type ID struct {
	Project string `json:"project"`
	Name    string `json:"part_name"`
	Number  uint8  `json:"part_number"`
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s-%d", id.Project, id.Name, id.Number)
}
