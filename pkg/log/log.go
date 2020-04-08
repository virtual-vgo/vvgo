package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

func Logger() *logrus.Logger { return logger }

var logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	},
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.InfoLevel,
	ExitFunc:     os.Exit,
	ReportCaller: false,
}
