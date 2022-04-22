package container

import (
	"io"

	"github.com/sirupsen/logrus"
)

func SetLogToFile(w io.Writer) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(w)
}

type StdoutRedirect struct{}

func (s *StdoutRedirect) Write(data []byte) (int, error) {
	logrus.WithFields(logrus.Fields{
		"stream": "stdout",
	}).Info(string(data))
	return len(data), nil
}

type StderrRedirect struct{}

func (s *StderrRedirect) Write(data []byte) (int, error) {
	logrus.WithFields(logrus.Fields{
		"stream": "stderr",
	}).Info(string(data))
	return len(data), nil
}
