package core

import "time"

type LogLevel int

const (
	Trace LogLevel = iota * 10
	Debug
	Info
	Warning
	Error
	Exception
)

func (l LogLevel) String() string {
	switch l {
	case Trace:
		return "TRACE"
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case Exception:
		return "EXCEPTION"
	default:
		return "UNKNOWN"
	}
}

type LogRecord struct {
	Level     LogLevel
	Timestamp time.Time
	Message   string
	Fields    map[string]interface{}

	Caller string
}
