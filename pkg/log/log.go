package log

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

func (x Logger) WithFields(fields logrus.Fields) Entry {
	return Entry{Entry: x.Logger.WithFields(fields)}
}

func (x Logger) JsonDecodeFailure(ctx context.Context, err error) {
	x.WithContext(ctx).SomeMethodFailure(ctx, "json.Decode", err)
}

func (x Logger) SomeMethodFailure(ctx context.Context, method string, err error) {
	x.WithContext(ctx).SomeMethodFailure(ctx, method, err)
}

func (x Logger) WithContext(ctx context.Context) Entry {
	return Entry{Entry: x.Logger.WithContext(ctx)}
}

func (e Entry) WithFields(fields logrus.Fields) Entry {
	return Entry{Entry: e.Entry.WithFields(fields)}
}

func (e Entry) WithError(err error) Entry {
	return Entry{Entry: e.Entry.WithError(err)}
}

func (e Entry) JsonDecodeFailure(ctx context.Context, err error) {
	e.SomeMethodFailure(ctx, "json.Decode", err)
}

func (e Entry) SomeMethodFailure(ctx context.Context, method string, err error) {
	e.WithContext(ctx).WithError(err).Error(method + "() failed")
}
