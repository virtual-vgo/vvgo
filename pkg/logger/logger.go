//go:generate go run ../../cmd/generate_code pkg/logger

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

type Logger struct{ *logrus.Logger }

func New() Logger { return defaultLogger }

type Entry struct{ *logrus.Entry }

func NewEntry() Entry            { return Entry{logrus.NewEntry(defaultLogger.Logger)} }
func (x Logger) NewEntry() Entry { return Entry{logrus.NewEntry(defaultLogger.Logger)} }

func StdLogger() *log.Logger            { return log.New(defaultLogger.Writer(), "", 0) }
func (x Logger) StdLogger() *log.Logger { return log.New(x.Writer(), "", 0) }
func (e Entry) StdLogger() *log.Logger  { return log.New(e.Writer(), "", 0) }

func MethodFailure(ctx context.Context, method string, err error) {
	defaultLogger.MethodFailure(ctx, method, err)
}
func (x Logger) MethodFailure(ctx context.Context, method string, err error) {
	x.NewEntry().MethodFailure(ctx, method, err)
}
func (e Entry) MethodFailure(ctx context.Context, method string, err error) {
	e.WithContext(ctx).WithError(err).Error(method + "() failed")
}
