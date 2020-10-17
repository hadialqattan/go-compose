/*
Logging utility.
*/

package utils

import (
	"io"

	"github.com/sirupsen/logrus"
)

type logger struct {
	writer io.Writer
}

func (log *logger) Write(msg []byte) (int, error) {
	return log.writer.Write(msg)
}

func newCustomLogger() *logrus.Logger {
	logger := logrus.New()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logger.SetFormatter(customFormatter)
	return logger
}
