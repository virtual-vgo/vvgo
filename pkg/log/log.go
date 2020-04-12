package log

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

var logger = &logrus.Logger{
	Out:          os.Stdout,
	Formatter:    new(logrus.JSONFormatter),
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.InfoLevel,
	ExitFunc:     os.Exit,
	ReportCaller: true,
}

func Logger() *logrus.Logger { return logger }

func StdLogger() *log.Logger {
	return log.New(logger.Writer(), "", 0)
}
