/*
Logging utility.
*/

package utils

import (
	"io"
)

type logger struct {
	writer io.Writer
}

func (log *logger) Write(msg []byte) (int, error) {
	return log.writer.Write(msg)
}
