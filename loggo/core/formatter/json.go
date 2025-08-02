package formatter

import (
	"bytes"
	"funchooooza-ossh/loggo/core"
	"strconv"
	"strings"
	"time"
)

// JsonFormatter сериализует LogRecord в JSON-подобный формат без зависимостей.
type JsonFormatter struct {
	style *core.FormatStyle
}

// NewJsonFormatter создаёт JsonFormatter с заданным стилем (или дефолтным).
func NewJsonFormatter(style *core.FormatStyle) *JsonFormatter {
	if style == nil {
		style = &core.FormatStyle{
			ColorKeys:   false,
			ColorValues: false,
			ColorLevel:  false,
			KeyColor:    "\033[36m", // голубой
			ValueColor:  "\033[37m", // белый/серый
			Reset:       "\033[0m",
		}
	}
	return &JsonFormatter{style: style}
}

// Format преобразует LogRecord в JSON-байты.
func (f *JsonFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')

	// "level":"INFO" (с цветом уровня)
	b.WriteString(`"`)
	b.WriteString(f.colorizeKey("level"))
	b.WriteString(`":"`)
	if f.style.ColorLevel {
		b.WriteString(r.Level.Color())
	}
	b.WriteString(r.Level.String())
	if f.style.ColorLevel {
		b.WriteString(f.style.Reset)
	}
	b.WriteByte('"')

	// ,"ts":"..."
	b.WriteString(`,"`)
	b.WriteString(f.colorizeKey("ts"))
	b.WriteString(`":"`)
	b.WriteString(r.Timestamp.Format(time.RFC3339Nano))
	b.WriteByte('"')

	// ,"msg":"..."
	b.WriteString(`,"`)
	b.WriteString(f.colorizeKey("msg"))
	b.WriteString(`":"`)
	b.WriteString(f.colorizeValue(escapeString(r.Message)))
	b.WriteByte('"')

	// поля из Fields
	for k, v := range r.Fields {
		b.WriteString(`,"`)
		b.WriteString(f.colorizeKey(escapeString(k)))
		b.WriteString(`":`)
		f.writeValue(&b, v)
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

// writeValue пишет значение в json-буфер в зависимости от типа.
func (f *JsonFormatter) writeValue(b *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		b.WriteByte('"')
		b.WriteString(f.colorizeValue(escapeString(val)))
		b.WriteByte('"')
	case int, int32, int64:
		b.WriteString(f.colorizeValue(toIntString(val)))
	case float64, float32:
		b.WriteString(f.colorizeValue(toFloatString(val)))
	case bool:
		b.WriteString(f.colorizeValue(strconv.FormatBool(val)))
	default:
		b.WriteString(`"unsupported_type"`)
	}
}

// colorizeKey возвращает ключ с ANSI-цветом, если включено.
func (f *JsonFormatter) colorizeKey(key string) string {
	if f.style.ColorKeys {
		return f.style.KeyColor + key + f.style.Reset
	}
	return key
}

// colorizeValue возвращает значение с ANSI-цветом, если включено.
func (f *JsonFormatter) colorizeValue(val string) string {
	if f.style.ColorValues {
		return f.style.ValueColor + val + f.style.Reset
	}
	return val
}

// escapeString экранирует кавычки и обратные слеши.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
