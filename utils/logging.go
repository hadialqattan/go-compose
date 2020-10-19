/*
Logging utility.
*/

package utils

import (
	"io"

	nested "github.com/antonfisher/nested-logrus-formatter"
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
	logger.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{"prefix"},
		TimestampFormat: "2006-01-02 15:04:05.00",
	})
	return logger
}
