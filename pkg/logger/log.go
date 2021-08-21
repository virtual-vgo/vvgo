//go:generate go run ../../tools/generate_code pkg/logger generated_methods.go

package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

var defaultLogger = Logger{
	Logger: &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	},
}

type Entry struct {
	*logrus.Entry
}

type Logger struct {
	*logrus.Logger
}

func New() Logger { return defaultLogger }

func StdLogger() *log.Logger {
	return log.New(defaultLogger.Writer(), "", 0)
}

func MethodFailure(ctx context.Context, method string, err error) {
	defaultLogger.MethodFailure(ctx, method, err)
}

func (x Logger) MethodFailure(ctx context.Context, method string, err error) {
	x.WithContext(ctx).MethodFailure(ctx, method, err)
}

func (e Entry) MethodFailure(ctx context.Context, method string, err error) {
	e.WithContext(ctx).WithError(err).Error(method + "() failed")
}
