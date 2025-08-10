package core

import (
	"context"
	"sync"
	"time"
)

type RouteProcessor struct {
	Formatter      FormatProcessor
	Writer         WriteProcessor
	LevelThreshold LogLevel

	queue  chan LogRecordRaw
	closed bool
	mu     sync.RWMutex
}

func NewRouteProcessor(formatter FormatProcessor, writer WriteProcessor, level LogLevel) *RouteProcessor {
	return &RouteProcessor{
		Formatter:      formatter,
		Writer:         writer,
		LevelThreshold: level,
		queue:          make(chan LogRecordRaw, 1024),
	}
}

func (r *RouteProcessor) ShouldLog(level LogLevel) bool {
	return level >= r.LevelThreshold
}

// safe put record into queue, delevery guaranteed
func (r *RouteProcessor) Enqueue(record LogRecordRaw) {
	r.mu.RLock() // lock
	defer r.mu.RUnlock()
	if r.closed { //check closed
		return
	}
	r.queue <- record //delivery asap
}

// reading for queue
func (r *RouteProcessor) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1) //can wait until complete from outside
	go func() {
		defer wg.Done()      //guarantee complete signal
		defer r.drainQueue() //guarantee to process every record in queue before close

		for {
			select {
			case rec, ok := <-r.queue:
				if !ok {
					return // if queue closed -> exit
				}
				record := rawToRecord(rec)
				if data, err := r.Formatter.Format(record); err == nil {
					_ = r.Writer.Write(data)
				}
			case <-ctx.Done():
				// w8ing queue to close, defer drain will finish processing
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

		for i := range b {
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

// delivery guarantee
func (r *RouteProcessor) drainQueue() {
	for rec := range r.queue { //queue range will work until chan not closed from outside
		record := rawToRecord(rec)
		if data, err := r.Formatter.Format(record); err == nil {
			_ = r.Writer.Write(data)
		}
	}

	if f, ok := r.Writer.(FlushableWriter); ok {
		_ = f.Flush()
	}
}

func (r *RouteProcessor) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed { // if closed -> return
		return
	}

	close(r.queue)  // else close
	r.closed = true //set flag
}
