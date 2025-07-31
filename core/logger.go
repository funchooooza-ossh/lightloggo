package core

import (
	"runtime"
	"strconv"
	"time"
)

type Logger struct {
	Routes []RouteProcessor
}

func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	// Получить caller
	_, file, line, ok := runtime.Caller(2)
	var caller string
	if ok {
		caller = file + ":" + itoa(line)
	}

	record := LogRecord{
		Level:     level,
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
		Caller:    caller,
	}

	for _, route := range l.Routes {
		_ = route.Process(record) // errors можно логировать позже
	}
}

// Упрощённые sugar-методы
func (l *Logger) Trace(msg string, fields map[string]interface{}) { l.log(Trace, msg, fields) }
func (l *Logger) Debug(msg string, fields map[string]interface{}) { l.log(Debug, msg, fields) }
func (l *Logger) Info(msg string, fields map[string]interface{})  { l.log(Info, msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]interface{})  { l.log(Warning, msg, fields) }
func (l *Logger) Error(msg string, fields map[string]interface{}) { l.log(Error, msg, fields) }
func (l *Logger) Exception(msg string, fields map[string]interface{}) {
	l.log(Exception, msg, fields)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
