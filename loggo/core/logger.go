package core

import (
	"context"
	"sync"
)

// Logger управляет маршрутизацией логов и жизненным циклом воркеров.
type Logger struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	routes []*RouteProcessor
}

// NewLogger создаёт асинхронный логгер с переданными маршрутизаторами.
func NewLogger(routes ...*RouteProcessor) *Logger {
	ctx, cancel := context.WithCancel(context.Background())

	logger := &Logger{
		ctx:    ctx,
		cancel: cancel,
		routes: routes,
	}

	for _, r := range routes {
		r.Start(ctx, &logger.wg)
	}

	return logger
}

// Close корректно завершает все воркеры, дожидаясь полной обработки очередей и вызова Flush().
func (l *Logger) Close() {
	for _, r := range l.routes {
		r.Close()
	}
	l.cancel()
	l.wg.Wait()
}
