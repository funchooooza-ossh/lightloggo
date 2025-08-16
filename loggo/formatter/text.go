// Package formatter provides concrete implementations of the core.FormatProcessor
// interface, such as formatters for plain text and JSON.
package formatter

import (
	"bytes"
	"fmt"
	"funchooooza-ossh/loggo/core"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TextFormatter serializes a LogRecord into a human-readable, single-line text format.
// It supports custom styling (colors) and deep rendering of structured data.
type TextFormatter struct {
	style    *core.FormatStyle
	MaxDepth int
}

// NewTextFormatter is the constructor for TextFormatter.
// If a nil style is provided, a default, non-colored style is used.
// If a nil maxDepth is provided, a default depth is used.
func NewTextFormatter(style *core.FormatStyle, maxDepth *int) *TextFormatter {
	var depth int
	if maxDepth == nil {
		depth = defaultDepth
	} else {
		depth = *maxDepth
	}

	if style == nil {
		style = core.NewDefaultStyle()
	}

	return &TextFormatter{style: style, MaxDepth: depth}
}

// Format takes a LogRecord and transforms it into a formatted byte slice.
// This implementation is guaranteed to not return an error.
func (f *TextFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer

	// 1. Add timestamp in a fixed format.
	// Example: [2025-08-14 15:30:00.000]
	b.WriteString("[")
	b.WriteString(r.Timestamp.Format("2006-01-02 15:04:05.000"))
	b.WriteString("] ")

	// 2. Add the log level, padded for alignment, with optional color.
	if f.style.ColorLevel {
		b.WriteString(r.Level.Color())
	}
	b.WriteString(padLevel(r.Level.String()))
	if f.style.ColorLevel {
		b.WriteString(f.style.Reset)
	}
	b.WriteByte(' ')

	// 3. Add the main log message.
	b.WriteString("→ ")
	b.WriteString(r.Message)

	// 4. Add structured fields if they exist.
	if len(r.Fields) > 0 {
		b.WriteString(" |")
		// Sort keys to ensure a stable, deterministic output order.
		keys := make([]string, 0, len(r.Fields))
		for k := range r.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// A map to track visited pointers to prevent infinite loops in cyclic data.
		visited := make(map[uintptr]struct{})
		for _, k := range keys {
			b.WriteByte(' ')
			b.WriteString(f.colorizeKey(k))
			b.WriteByte('=')
			// Delegate the complex task of value rendering to the renderText function.
			f.renderText(&b, r.Fields[k], 0, visited)
		}
	}

	return b.Bytes(), nil
}

// renderText is the recursive core of the formatter. It inspects the type of `v`
// and writes its string representation into the buffer `b`.
// It handles simple types directly and delegates complex ones to specialized methods.
func (f *TextFormatter) renderText(b *bytes.Buffer, v any, depth int, visited map[uintptr]struct{}) {
	// --- Base cases for recursion termination ---
	if depth >= f.MaxDepth {
		b.WriteString(f.colorizeValue("<max_depth>"))
		return
	}
	if d, ok := v.(time.Duration); ok {
		b.WriteString(f.colorizeValue(d.String()))
		return
	}

	// --- Fast path for common, simple types ---
	switch x := v.(type) {
	case nil:
		b.WriteString(f.colorizeValue("null"))
	case string:
		s := addMultilinePrefix(x)
		b.WriteString(f.colorizeValue(strconv.Quote(s)))
	case bool:
		b.WriteString(f.colorizeValue(strconv.FormatBool(x)))
	case int, int8, int16, int32, int64:
		b.WriteString(f.colorizeValue(strconv.FormatInt(reflect.ValueOf(x).Int(), 10)))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		b.WriteString(f.colorizeValue(strconv.FormatUint(reflect.ValueOf(x).Uint(), 10)))
	case float32, float64:
		b.WriteString(f.colorizeValue(toFloatString(x)))

	// --- Slower path using reflection for complex types ---
	default:
		f.renderByReflect(b, v, depth, visited)
	}
}

func (f *TextFormatter) renderByReflect(b *bytes.Buffer, v any, depth int, visited map[uintptr]struct{}) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		b.WriteString(f.colorizeValue("null"))
		return
	}

	if ok, release := markAndCheck(rv, visited); !ok {
		b.WriteString(f.colorizeValue("<cycle>"))
		return
	} else {
		defer release()
	}

	switch rv.Kind() {

	case reflect.Interface:
		if rv.IsNil() {
			b.WriteString(f.colorizeValue("null"))
			return
		}

		f.renderText(b, rv.Elem().Interface(), depth+1, visited)

	// --- DELEGATION TO HELPER METHODS ---
	case reflect.Ptr:
		f.renderPtr(b, rv, depth, visited)
	case reflect.Struct:
		f.renderStruct(b, rv, depth, visited)
	case reflect.Map:
		f.renderMap(b, rv, depth, visited)
	case reflect.Slice, reflect.Array:
		f.renderSlice(b, rv, depth, visited)

	default:
		// Fallback for any other unhandled primitive-like type.
		b.WriteString(f.colorizeValue(fmt.Sprint(v)))
	}
}

// --- Private helper methods for rendering ---

// colorizeKey applies color to a key string if style.ColorKeys is enabled.
func (f *TextFormatter) colorizeKey(k string) string {
	if f.style.ColorKeys {
		return f.style.KeyColor + k + f.style.Reset
	}
	return k
}

// colorizeValue applies color to a value string if style.ColorValues is enabled.
func (f *TextFormatter) colorizeValue(v string) string {
	if f.style.ColorValues {
		return f.style.ValueColor + v + f.style.Reset
	}
	return v
}

// padLevel adds padding to a log level string for consistent alignment.
func padLevel(level string) string {
	const minWidth = 7
	if len(level) < minWidth {
		return level + strings.Repeat(" ", minWidth-len(level))
	}
	return level
}

func (f *TextFormatter) renderPtr(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if rv.IsNil() {
		b.WriteString(f.colorizeValue("null"))
		return
	}

	f.renderText(b, rv.Elem().Interface(), depth+1, visited)

}

// renderStruct handles the reflection-based rendering of struct types.
// It respects `json` tags for field naming, skipping, and omitempty options.
// Fields are sorted alphabetically by their effective key for stable output.
func (f *TextFormatter) renderStruct(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	type kv struct {
		key string
		idx int
	}
	t := rv.Type()
	fields := make([]kv, 0, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // Skip unexported fields
			continue
		}
		key := sf.Name
		if tag := sf.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				key = parts[0]
			}
			for _, opt := range parts[1:] {
				if opt == "omitempty" && rv.Field(i).IsZero() {
					key = "" // Mark for skipping
					break
				}
			}
			if key == "" {
				continue
			}
		}
		fields = append(fields, kv{key, i})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })

	b.WriteByte('{')
	for i, fdef := range fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.colorizeKey(fdef.key))
		b.WriteString(": ")
		f.renderText(b, rv.Field(fdef.idx).Interface(), depth+1, visited)
	}
	b.WriteByte('}')
}

// renderMap handles the reflection-based rendering of map types.
// It sorts maps by their keys to ensure a stable, deterministic output.
// Only maps with string keys are fully rendered; others are marked as unsupported.
func (f *TextFormatter) renderMap(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if rv.Type().Key().Kind() != reflect.String {
		b.WriteString(f.colorizeValue("<unsupported_map_key>"))
		return
	}
	keys := rv.MapKeys()
	// Sort keys for stable output
	ss := make([]string, len(keys))
	for i, k := range keys {
		ss[i] = k.String()
	}
	sort.Strings(ss)

	b.WriteByte('{')
	for i, k := range ss {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.colorizeKey(k))
		b.WriteByte(':')
		b.WriteByte(' ')
		f.renderText(b, rv.MapIndex(reflect.ValueOf(k)).Interface(), depth+1, visited)
	}
	b.WriteByte('}')
}

// renderSlice handles the reflection-based rendering of slice and array types.
// It provides special handling for []byte for a more concise output.
func (f *TextFormatter) renderSlice(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if rv.Type().Elem().Kind() == reflect.Uint8 {
		b.WriteString(f.colorizeValue(fmt.Sprintf("[]byte(%d)", rv.Len())))
		return
	}
	n := rv.Len()
	b.WriteByte('[')
	for i := range n {
		if i > 0 {
			b.WriteString(", ")
		}
		f.renderText(b, rv.Index(i).Interface(), depth+1, visited)
	}
	b.WriteByte(']')
}
