package writer

import (
	"os"
)

// StdoutWriter пишет логи в стандартный вывод.
type StdoutWriter struct{}

// NewStdoutWriter создаёт StdoutWriter.
func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{}
}

// Write выводит отформатированные данные в stdout, добавляя перенос строки.
func (w *StdoutWriter) Write(data []byte) error {
	_, err := os.Stdout.Write(append(data, '\n'))
	return err
}

// Flush реализует интерфейс Flushable, но ничего не делает (stdout не буферизуется).
func (w *StdoutWriter) Flush() error {
	return nil
}
