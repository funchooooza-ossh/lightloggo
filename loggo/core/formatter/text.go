package formatter

import (
	"bytes"
	"funchooooza-ossh/loggo/core"
	"sort"
	"strconv"
	"strings"
)

type TextFormatter struct {
	style *core.FormatStyle
}

func NewTextFormatter(style *core.FormatStyle) *TextFormatter {
	if style == nil {
		style = &core.FormatStyle{
			ColorKeys:   false,
			ColorValues: false,
			ColorLevel:  false,
			KeyColor:    "\033[36m", // голубой
			ValueColor:  "\033[37m",
			Reset:       "\033[0m",
		}
	}
	return &TextFormatter{style: style}
}

func (f *TextFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer

	// [timestamp]
	b.WriteString("[")
	b.WriteString(r.Timestamp.Format("2006-01-02 15:04:05.000"))
	b.WriteString("] ")

	// LEVEL
	if f.style.ColorLevel {
		b.WriteString(r.Level.Color())
	}
	b.WriteString(padLevel(r.Level.String()))
	if f.style.ColorLevel {
		b.WriteString(f.style.Reset)
	}
	b.WriteByte(' ')

	// caller
	if r.Caller != "" {
		b.WriteString(r.Caller)
		b.WriteByte(' ')
	}

	// → message
	b.WriteString("→ ")
	b.WriteString(r.Message)

	// поля (отсортированы для стабильности)
	if len(r.Fields) > 0 {
		b.WriteString(" |")
		keys := make([]string, 0, len(r.Fields))
		for k := range r.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := r.Fields[k]
			b.WriteByte(' ')
			b.WriteString(f.colorizeKey(k))
			b.WriteByte('=')
			b.WriteString(f.colorizeValue(toString(v)))
		}
	}

	return b.Bytes(), nil
}

func (f *TextFormatter) colorizeKey(k string) string {
	if f.style.ColorKeys {
		return f.style.KeyColor + k + f.style.Reset
	}
	return k
}

func (f *TextFormatter) colorizeValue(v string) string {
	if f.style.ColorValues {
		return f.style.ValueColor + v + f.style.Reset
	}
	return v
}

func padLevel(level string) string {
	if len(level) < 7 {
		return level + strings.Repeat(" ", 7-len(level))
	}
	return level
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int32, int64:
		return toIntString(val)
	case float64, float32:
		return toFloatString(val)
	case bool:
		return strconv.FormatBool(val)
	default:
		return "<unsupported>"
	}
}
