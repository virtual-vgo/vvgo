//go:generate go run ../../tools/generate_code pkg/errors

package errors

import "errors"

func Is(err error, target error) bool       { return errors.Is(err, target) }
func As(err error, target interface{}) bool { return errors.As(err, target) }
func Unwrap(err error) error                { return errors.Unwrap(err) }
func New(text string) error                 { return errors.New(text) }

func Non200StatusCode() error { return errors.New("non-200 status code") }
