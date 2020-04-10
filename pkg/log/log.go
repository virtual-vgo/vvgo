package log

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

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

func Logger() *logrus.Logger { return logger }

func StdLogger() *log.Logger {
	return log.New(logger.Writer(), "", 0)
}
