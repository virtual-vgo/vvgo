//go:generate go run ../../cmd/generate_code pkg/logger

package logger

import (
	"context"
	"github.com/sirupsen/logrus"
)

var Logger = logrus.StandardLogger()

type Entry struct{ *logrus.Entry }

func MethodFailure(ctx context.Context, method string, err error) {
	logrus.NewEntry(Logger).WithContext(ctx).WithError(err).Error(method + "() failed")
}
func (e Entry) MethodFailure(ctx context.Context, method string, err error) {
	e.WithContext(ctx).WithError(err).Error(method + "() failed")
}
