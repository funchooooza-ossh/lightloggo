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

func (lvl LogLevel) Color() string {
	switch lvl {
	case Trace:
		return "\033[90m" // серый
	case Debug:
		return "\033[34m" // синий
	case Info:
		return "\033[32m" // зелёный
	case Warning:
		return "\033[33m" // жёлтый
	case Error:
		return "\033[31m" // красный
	case Exception:
		return "\033[1;31m" // ярко-красный
	default:
		return "\033[0m"
	}
}

func (lvl LogLevel) Reset() string {
	return "\033[0m"
}

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
}

type LogRecordRaw struct {
	Level   LogLevel
	Message []byte
	Fields  []byte
}
