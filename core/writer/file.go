package writer

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
)

type FileWriter struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

// NewFileWriter создает FileWriter с буферизацией.
func NewFileWriter(path string) (*FileWriter, error) {
	// Создаём директории, если не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// Открываем файл (append mode, create if not exist)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		file:   f,
		writer: bufio.NewWriterSize(f, 4096), // 4 KB буфер
	}, nil
}

// Write реализует WriteProcessor.
func (fw *FileWriter) Write(data []byte) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	_, err := fw.writer.Write(append(data, '\n'))
	return err
}

// Flush реализует FlushableWriteProcessor.
func (fw *FileWriter) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.writer.Flush()
}

// Close закрывает файл и сбрасывает буфер.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	_ = fw.writer.Flush()
	return fw.file.Close()
}
