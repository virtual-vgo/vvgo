package error_wrappers

import (
	"errors"
	"fmt"
)

func NewRequestFailed(err error) error {
	return fmt.Errorf("http.NewRequest() failed: %w", err)
}

func HTTPDoFailed(err error) error {
	return fmt.Errorf("http.Do() failed: %w", err)
}

func Non200StatusCode() error {
	return errors.New("non-200 status code")
}
