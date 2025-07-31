package writer

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileWriter struct {
	path       string
	maxSizeMB  int64
	maxBackups int
	compress   bool

	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
	size   int64
}

// NewFileWriter создаёт новый лог-файл с опциями ротации и сжатия.
func NewFileWriter(path string, maxSizeMB int64, maxBackups int, compress bool) (*FileWriter, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()

	return &FileWriter{
		path:       path,
		maxSizeMB:  maxSizeMB,
		maxBackups: maxBackups,
		compress:   compress,
		file:       f,
		writer:     bufio.NewWriter(f),
		size:       info.Size(),
	}, nil
}

func (fw *FileWriter) Write(p []byte) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.shouldRotate(len(p)) {
		if err := fw.rotate(); err != nil {
			return err
		}
	}

	n, err := fw.writer.Write(append(p, '\n'))
	fw.size += int64(n)
	return err
}

func (fw *FileWriter) Flush() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.writer.Flush()
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	_ = fw.writer.Flush()
	return fw.file.Close()
}

// --- rotation logic ---

func (fw *FileWriter) shouldRotate(incoming int) bool {
	return fw.maxSizeMB > 0 && fw.size+int64(incoming) > fw.maxSizeMB*1024*1024
}

func (fw *FileWriter) rotate() error {
	fw.writer.Flush()
	fw.file.Close()

	timestamp := time.Now().Format("2006-01-02T15-04-05")
	rotatedName := fw.path + "." + timestamp
	if err := os.Rename(fw.path, rotatedName); err != nil {
		return err
	}

	if fw.compress {
		go compressFile(rotatedName)
	}

	f, err := os.OpenFile(fw.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	fw.file = f
	fw.writer = bufio.NewWriter(f)
	fw.size = 0

	fw.cleanupBackups()

	return nil
}

func (fw *FileWriter) cleanupBackups() {
	if fw.maxBackups <= 0 {
		return
	}

	dir := filepath.Dir(fw.path)
	prefix := filepath.Base(fw.path)

	files, _ := os.ReadDir(dir)
	var backups []string

	for _, f := range files {
		name := f.Name()

		// Ищем только архивы, начинающиеся с app.log. и заканчивающиеся на .gz
		if strings.HasPrefix(name, prefix+".") && strings.HasSuffix(name, ".gz") {
			backups = append(backups, filepath.Join(dir, name))
		}
	}

	if len(backups) <= fw.maxBackups {
		return
	}

	// Сортируем по имени — время встроено в имя
	sort.Strings(backups)

	// Удаляем лишние архивы
	for _, f := range backups[:len(backups)-fw.maxBackups] {
		_ = os.Remove(f)
	}
}

func compressFile(path string) {
	in, err := os.Open(path)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(path + ".gz")
	if err != nil {
		return
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	io.Copy(gz, in)
	gz.Close()

	os.Remove(path)
}
