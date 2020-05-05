package parts

import (
	"bytes"
	"context"
	"encoding"
	"encoding/gob"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"strings"
	"time"
)

const DataKey = "parts"

var (
	ErrInvalidPartName   = fmt.Errorf("invalid part name")
	ErrInvalidPartNumber = fmt.Errorf("invalid part number")
)

type Hash interface {
	HVals(ctx context.Context) ([][]byte, error)
	HGet(ctx context.Context, name string, dest encoding.BinaryUnmarshaler) error
	HSet(ctx context.Context, name string, src encoding.BinaryMarshaler) error
}

type Locker interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context)
}

type Parts struct {
	Hash
	Locker
}

func (x Parts) List(ctx context.Context) ([]Part, error) {
	vals, err := x.Hash.HVals(ctx)
	if err != nil {
		return nil, err
	}
	parts := make([]Part, len(vals))
	for i := range vals {
		if err := parts[i].UnmarshalBinary(vals[i]); err != nil {
			return nil, fmt.Errorf("part.UnmarshalBinary() failed: %w", err)
		}
	}
	return parts, nil
}

func (x Parts) Save(ctx context.Context, parts []Part) error {
	// first, validate all the parts
	if err := validatePart(parts); err != nil {
		return err
	}

	// we'll be making quite a few changes at once
	if err := x.Lock(ctx); err != nil {
		return err
	}
	defer x.Unlock(ctx)

	// read the existing parts
	cur := make([]Part, len(parts))
	for i := range parts {
		if err := x.Hash.HGet(ctx, parts[i].ID.String(), &cur[i]); err != storage.ErrKeyIsEmpty && err != nil {
			return err
		}
	}

	// merge
	for i := range parts {
		// append the old fields
		parts[i].Sheets = append(parts[i].Sheets, cur[i].Sheets...)
		parts[i].Clix = append(parts[i].Clix, cur[i].Clix...)
	}

	// write changes
	for _, part := range parts {
		if err := x.Hash.HSet(ctx, part.ID.String(), &part); err != nil {
			return err
		}
	}
	return nil
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
	Clix   Links
	Sheets Links
}

type Data Part

func (x *Part) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode((*Data)(x))
	return buf.Bytes(), err
}

func (x *Part) UnmarshalBinary(src []byte) error {
	return gob.NewDecoder(bytes.NewReader(src)).Decode((*Data)(x))
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
