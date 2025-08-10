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

type TextFormatter struct {
	style    *core.FormatStyle
	MaxDepth int
}

func NewTextFormatter(style *core.FormatStyle, maxDepth *int) *TextFormatter {
	var depth int
	if maxDepth == nil {
		depth = defaultDepth
	} else {
		depth = *maxDepth
	}
	if style == nil {
		style = &core.FormatStyle{
			ColorKeys:   false,
			ColorValues: false,
			ColorLevel:  false,
			KeyColor:    "\033[36m",
			ValueColor:  "\033[37m",
			Reset:       "\033[0m",
		}
	}
	return &TextFormatter{style: style, MaxDepth: depth}
}

func (f *TextFormatter) sep() string {
	return " | "
}
func (f *TextFormatter) arrow() string {
	return " -> "
}

func (f *TextFormatter) writeStyledLevel(b *bytes.Buffer, lvl core.LogLevel) {
	if f.style.ColorLevel {
		b.WriteString(lvl.Color())
	}
	b.WriteString(lvl.String())
	if f.style.ColorLevel {
		b.WriteString(f.style.Reset)
	}
}

func (f *TextFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer

	visited := getVisited()
	defer putVisited(visited)

	// [timestamp]
	b.WriteString(r.Timestamp.Format("2006-01-02 15:04:05.000"))

	// LEVEL
	b.WriteString(f.sep())
	f.writeStyledLevel(&b, r.Level)

	// → message
	b.WriteString(f.sep())
	b.WriteString(f.arrow())
	b.WriteString(r.Message)

	// sorted fields
	if len(r.Fields) > 0 {
		b.WriteString(" |")
		keys := make([]string, 0, len(r.Fields))
		for k := range r.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteByte(' ')
			b.WriteString(f.colorizeKey(k))
			b.WriteByte('=')
			f.renderText(&b, r.Fields[k], 0, visited)
		}
	}
	return b.Bytes(), nil
}

func (f *TextFormatter) renderText(b *bytes.Buffer, v any, depth int, visited map[uintptr]struct{}) {
	if depth >= f.MaxDepth {
		b.WriteString(f.colorizeValue("<max_depth>")) //depth protect
	}

	if d, ok := v.(time.Duration); ok {
		b.WriteString(f.colorizeValue(d.String())) // non reflect type
		return
	}

	switch x := v.(type) {
	case nil:
		b.WriteString(f.colorizeValue("null")) // non reflect type

	case string:
		s := addMultilinePrefix(x)
		b.WriteString(f.colorizeValue(strconv.Quote(s)))

	case bool:
		if x {
			b.WriteString(f.colorizeValue("true")) // non reflect scalar boolean
		} else {
			b.WriteString(f.colorizeValue("false"))
		}

	case int, int8, int16, int32, int64:
		b.WriteString(f.colorizeValue(strconv.FormatInt(reflect.ValueOf(x).Int(), 10)))

	case uint, uint8, uint16, uint32, uint64, uintptr:
		b.WriteString(f.colorizeValue(strconv.FormatUint(reflect.ValueOf(x).Uint(), 10)))

	case float32, float64:
		b.WriteString(f.colorizeValue(toFloatString(x)))

	case map[string]any:
		// cycling protection
		if ok, release := markAndCheck(reflect.ValueOf(x), visited); !ok {
			b.WriteString(f.colorizeValue("<cycle>"))
			return
		} else {
			defer release()
		}

		b.WriteByte('{')
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(f.colorizeKey(k))
			b.WriteString(": ")
			f.renderText(b, x[k], depth+1, visited)
		}
		b.WriteByte('}')

	case []any:
		// cycle protection
		if ok, release := markAndCheck(reflect.ValueOf(x), visited); !ok {
			b.WriteString(f.colorizeValue("<cycle>"))
			return
		} else {
			defer release()
		}

		b.WriteByte('[')
		for i := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			f.renderText(b, x[i], depth+1, visited)
		}
		b.WriteByte(']')

	default:
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			b.WriteString(f.colorizeValue("null"))
			return
		}

		// cycle protection
		if ok, release := markAndCheck(rv, visited); !ok {
			b.WriteString(f.colorizeValue("<cycle>"))
			return
		} else {
			defer release()
		}

		switch rv.Kind() {
		case reflect.Ptr:
			if rv.IsNil() {
				b.WriteString(f.colorizeValue("null"))
				return
			}
			f.renderText(b, rv.Elem().Interface(), depth+1, visited)

		case reflect.Interface:
			if rv.IsNil() {
				b.WriteString(f.colorizeValue("null"))
				return
			}
			f.renderText(b, rv.Elem().Interface(), depth+1, visited)

		case reflect.Struct:
			type kv struct {
				key string
				idx int
			}
			t := rv.Type()
			fields := make([]kv, 0, rv.NumField())
			for i := 0; i < rv.NumField(); i++ {
				sf := t.Field(i)
				if sf.PkgPath != "" {
					continue
				} // unexported
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
							key = ""
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

		case reflect.Map:
			// only string keys
			if rv.Type().Key().Kind() != reflect.String {
				b.WriteString(f.colorizeValue("<unsupported_map_key>"))
				return
			}
			keys := rv.MapKeys()
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
				b.WriteString(": ")
				f.renderText(b, rv.MapIndex(reflect.ValueOf(k)).Interface(), depth+1, visited)
			}
			b.WriteByte('}')

		case reflect.Slice, reflect.Array:
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				b.WriteString(f.colorizeValue(fmt.Sprintf("[]byte(%d)", rv.Len())))
				return
			}
			n := rv.Len()
			b.WriteByte('[')
			for i := 0; i < n; i++ {
				if i > 0 {
					b.WriteString(", ")
				}
				f.renderText(b, rv.Index(i).Interface(), depth+1, visited)
			}
			b.WriteByte(']')

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			b.WriteString(f.colorizeValue(strconv.FormatInt(rv.Int(), 10)))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			b.WriteString(f.colorizeValue(strconv.FormatUint(rv.Uint(), 10)))

		case reflect.Float32:
			b.WriteString(f.colorizeValue(strconv.FormatFloat(rv.Float(), 'f', -1, 32)))
		case reflect.Float64:
			b.WriteString(f.colorizeValue(strconv.FormatFloat(rv.Float(), 'f', -1, 64)))

		case reflect.Bool:
			if rv.Bool() {
				b.WriteString(f.colorizeValue("true"))
			} else {
				b.WriteString(f.colorizeValue("false"))
			}

		case reflect.String:
			s := addMultilinePrefix(rv.String())
			b.WriteString(f.colorizeValue(strconv.Quote(s)))

		default:
			b.WriteString(f.colorizeValue(fmt.Sprint(v)))
		}
	}
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
