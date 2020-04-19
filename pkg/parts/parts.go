package parts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"time"
)

const DataFile = "parts.json"

var (
	ErrInvalidPartName   = fmt.Errorf("invalid part name")
	ErrInvalidPartNumber = fmt.Errorf("invalid part number")
)

type Parts struct {
	*storage.Cache
	*locker.Locker
}

func (x Parts) Init(ctx context.Context) error {
	return x.PutObject(ctx, DataFile, storage.NewJSONObject(bytes.NewBuffer([]byte(`[]`))))
}

func (x Parts) List(ctx context.Context) ([]Part, error) {
	// grab the data file
	var dest storage.Object
	if err := x.GetObject(ctx, DataFile, &dest); err != nil {
		return nil, err
	}

	// deserialize the data file
	var parts []Part
	if err := json.NewDecoder(&dest.Buffer).Decode(&parts); err != nil {
		return nil, fmt.Errorf("json.Decode() failed: %v", err)
	}
	return parts, nil
}

func (x Parts) Save(ctx context.Context, parts []Part) error {
	// first, validate all the parts
	if err := validatePart(parts); err != nil {
		return err
	}

	// now we update the data file
	// read+modify+write means we need a lock
	if err := x.Lock(ctx); err != nil {
		return err
	}
	defer x.Unlock(ctx)

	// pull down the parts data
	allParts, err := x.List(ctx)
	if allParts == nil {
		return err
	}

	// merge the changes
	allParts = mergeChanges(allParts, parts)

	// encode the data file
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&allParts); err != nil {
		return fmt.Errorf("json.Encode() failed: %v", err)
	}

	return x.PutObject(ctx, DataFile, storage.NewJSONObject(&buffer))
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

func validatePart(parts []Part) error {
	for _, part := range parts {
		if err := part.Validate(); err != nil {
			return err
		}
	}
	return nil
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
