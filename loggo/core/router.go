package core

import (
	"context"
	"sync"
	"time"
)

// RouteProcessor связывает форматтер и writer, обрабатывает лог-события асинхронно.
type RouteProcessor struct {
	Formatter      FormatProcessor
	Writer         WriteProcessor
	LevelThreshold LogLevel

	queue  chan LogRecordRaw
	closed bool
	mu     sync.RWMutex
}

// NewRouteProcessor создаёт маршрутизатор логов с указанным форматтером и writer'ом.
func NewRouteProcessor(formatter FormatProcessor, writer WriteProcessor, level LogLevel) *RouteProcessor {
	return &RouteProcessor{
		Formatter:      formatter,
		Writer:         writer,
		LevelThreshold: level,
		queue:          make(chan LogRecordRaw, 1024),
	}
}

// ShouldLog проверяет, подходит ли уровень события для этого роута.
func (r *RouteProcessor) ShouldLog(level LogLevel) bool {
	return level >= r.LevelThreshold
}

// Enqueue отправляет событие в очередь логирования (если не закрыто).
func (r *RouteProcessor) Enqueue(record LogRecordRaw) {
	r.mu.RLock()
	closed := r.closed
	q := r.queue
	r.mu.RUnlock()
	if closed {
		return
	}
	q <- record
}

// Start запускает обработку очереди в отдельной горутине.
func (r *RouteProcessor) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer r.drainQueue()

		for {
			select {
			case rec, ok := <-r.queue:
				if !ok {
					return
				}
				record := rawToRecord(rec)
				if data, err := r.Formatter.Format(record); err == nil {
					_ = r.Writer.Write(data)
				}
			case <-ctx.Done():
				// просто ждём закрытия очереди, drain сделает остальное
				return
			}
		}
	}()
}

func rawToRecord(rec LogRecordRaw) LogRecord {
	fields := make(map[string]interface{})

	if len(rec.Fields) > 0 {
		b := rec.Fields
		start := 0
		var key string
		isKey := true

		for i := 0; i < len(b); i++ {
			if b[i] == 0 {
				part := string(b[start:i])
				if isKey {
					key = part
					isKey = false
				} else {
					fields[key] = part
					isKey = true
				}
				start = i + 1
			}
		}
	}

	var msg string
	if len(rec.Message) > 0 {
		msg = string(rec.Message)
	}

	return LogRecord{
		Level:     rec.Level,
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}
}

// drainQueue считывает остатки очереди и вызывает Flush().
func (r *RouteProcessor) drainQueue() {
	for rec := range r.queue {
		record := rawToRecord(rec)
		if data, err := r.Formatter.Format(record); err == nil {
			_ = r.Writer.Write(data)
		}
	}

	if f, ok := r.Writer.(FlushableWriter); ok {
		_ = f.Flush()
	}
}

// Close завершает работу: закрывает очередь (если ещё нет).
func (r *RouteProcessor) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	close(r.queue)
	r.closed = true
}
