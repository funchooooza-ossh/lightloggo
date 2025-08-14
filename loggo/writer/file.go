package writer

import (
	"bufio"
	"fmt"
	"funchooooza-ossh/loggo/compressor"
	"funchooooza-ossh/loggo/core"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type RotateInterval string

const (
	RotateDaily   RotateInterval = "day"
	RotateWeekly  RotateInterval = "week"
	RotateMonthly RotateInterval = "month"
)

type Compress string

const (
	Gz   Compress = "gz"
	Null Compress = ""
)

type FileWriter struct {
	path       string
	maxSizeMB  int64
	maxBackups int

	compress   Compress
	compressor core.Compressor

	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
	size   int64

	rotateInterval RotateInterval
	nextRotateTime time.Time
}

// NewFileWriter создаёт новый лог-файл с опциями ротации и сжатия.
func NewFileWriter(path string, maxSizeMB int64, maxBackups int, interval RotateInterval, compress *Compress) (*FileWriter, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	var comp core.Compressor
	compressVal := ""

	if compress != nil {
		switch *compress {
		case Gz:
			compressVal = "gz"
			comp = &compressor.GzipCompressor{}
		// можно добавить другие варианты позже
		default:
			return nil, fmt.Errorf("unsupported compression: %s", *compress)
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	info, statErr := f.Stat()
	if statErr != nil {
		f.Close()
		return nil, statErr
	}

	now := time.Now()
	return &FileWriter{
		path:           path,
		maxSizeMB:      maxSizeMB,
		maxBackups:     maxBackups,
		compress:       Compress(compressVal),
		compressor:     comp,
		file:           f,
		writer:         bufio.NewWriter(f),
		size:           info.Size(),
		rotateInterval: interval,
		nextRotateTime: nextRotation(now, interval),
	}, nil
}

func (fw *FileWriter) Write(p []byte) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.shouldRotateByTime(time.Now()) || fw.shouldRotateBySize(len(p)) {
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

func nextRotation(t time.Time, interval RotateInterval) time.Time {
	switch interval {
	case RotateDaily:
		return t.Truncate(24 * time.Hour).Add(24 * time.Hour)
	case RotateWeekly:
		offset := int(time.Monday - t.Weekday())
		if offset <= 0 {
			offset += 7
		}
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, offset)
	case RotateMonthly:
		return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Time{} // zero → no time rotation
	}
}

func (fw *FileWriter) shouldRotateByTime(now time.Time) bool {
	if fw.rotateInterval == "" {
		return false
	}
	return now.After(fw.nextRotateTime)
}

func (fw *FileWriter) shouldRotateBySize(incoming int) bool {
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

	if fw.compressor != nil {
		go func(src string) {
			dst := src + fw.compressor.Extension()
			_ = fw.compressor.Compress(src, dst)
			_ = os.Remove(src)
		}(rotatedName)
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
	prefix := filepath.Base(fw.path) + "."

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []string

	for _, f := range files {
		name := f.Name()

		// Ищем только те, что начинаются с basename+"."
		if strings.HasPrefix(name, prefix) {
			fullPath := filepath.Join(dir, name)
			backups = append(backups, fullPath)
		}
	}

	if len(backups) <= fw.maxBackups {
		return
	}

	// Сортируем по имени (в имени уже заложен timestamp)
	sort.Strings(backups)

	// Удаляем самые старые
	for _, f := range backups[:len(backups)-fw.maxBackups] {
		_ = os.Remove(f)
	}
}
