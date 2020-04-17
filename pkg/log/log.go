package log

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"os"
)

func Logger() *logrus.Logger {
	return &logrus.Logger{
		Out: io.MultiWriter(os.Stdout, discordWriter{
			endpoint: "",
		}),
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

// Probably want to use hooks here. Example logrus hook for syslog:
// https://github.com/sirupsen/logrus/blob/master/hooks/syslog/syslog.go

type discordWriter struct {
	endpoint string
}

func (x discordWriter) Write(msg []byte) (int, error) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&discordMessage{
		Content: string(msg),
	})
	resp, err := http.Post(x.endpoint, "application/json", &buf)
	if err != nil {
		Logger().WithError(err).Error("http.Post() failed")
	}
	if resp.StatusCode != http.StatusOK {
		Logger().WithField("status", resp.StatusCode).Error("non-200 status code")
	}
	return buf.Len(), nil
}

type discordMessage struct {
	Content string `json:"content"`
}
