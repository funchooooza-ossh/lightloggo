package formatter

import (
	"bytes"
	"funchooooza-ossh/loggo/core"
	"strconv"
	"strings"
	"time"
)

// JsonFormatter сериализует LogRecord в JSON-подобный формат без зависимостей.
type JsonFormatter struct{}

// NewJsonFormatter создаёт JsonFormatter.
func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{}
}

// Format преобразует LogRecord в JSON-байты.
func (f *JsonFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')

	// "level":"INFO"
	b.WriteString(`"level":"`)
	b.WriteString(r.Level.String())
	b.WriteByte('"')

	// ,"ts":"2025-07-31T12:00:00Z"
	b.WriteString(`,"ts":"`)
	b.WriteString(r.Timestamp.Format(time.RFC3339Nano))
	b.WriteByte('"')

	// ,"msg":"message text"
	b.WriteString(`,"msg":"`)
	b.WriteString(escapeString(r.Message))
	b.WriteByte('"')

	// ,"caller":"file.go:42"
	if r.Caller != "" {
		b.WriteString(`,"caller":"`)
		b.WriteString(escapeString(r.Caller))
		b.WriteByte('"')
	}

	// поля из Fields
	for k, v := range r.Fields {
		b.WriteByte(',')
		b.WriteByte('"')
		b.WriteString(escapeString(k))
		b.WriteString(`":`)
		writeValue(&b, v)
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

// writeValue пишет значение в json-буфер в зависимости от типа.
func writeValue(b *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		b.WriteByte('"')
		b.WriteString(escapeString(val))
		b.WriteByte('"')
	case int, int32, int64:
		b.WriteString(toIntString(val))
	case float64, float32:
		b.WriteString(toFloatString(val))
	case bool:
		b.WriteString(strconv.FormatBool(val))
	default:
		b.WriteString(`"unsupported_type"`)
	}
}

// escapeString экранирует кавычки и обратные слеши.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
