package writer

import (
	"os"
)

type StdoutWriter struct{}

func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{}
}

func (w *StdoutWriter) Write(data []byte) error {
	// Простой println с newline
	_, err := os.Stdout.Write(append(data, '\n'))
	return err
}
