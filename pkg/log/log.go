package log

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func Logger() *logrus.Logger {
	return &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

func StdLogger() *log.Logger {
	return log.New(Logger().Writer(), "", 0)
}
