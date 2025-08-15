package formatter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"funchooooza-ossh/loggo/core"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// JsonFormatter сериализует LogRecord в JSON-подобный формат без зависимостей.
type JsonFormatter struct {
	style    *core.FormatStyle
	MaxDepth int
}

// NewJsonFormatter создаёт JsonFormatter с заданным стилем (или дефолтным).
func NewJsonFormatter(style *core.FormatStyle, maxDepth *int) *JsonFormatter {
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
			KeyColor:    "\033[36m", // голубой
			ValueColor:  "\033[37m", // белый/серый
			Reset:       "\033[0m",
		}
	}
	return &JsonFormatter{style: style, MaxDepth: depth}
}

// Format преобразует LogRecord в JSON-байты.
func (f *JsonFormatter) Format(r core.LogRecord) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')

	//ANCHOR - LEVEL
	writeJSONString(&b, "level")
	b.WriteByte(':')
	writeJSONString(&b, r.Level.String())

	//ANCHOR - TIMESTAMP
	b.WriteByte(',')
	writeJSONString(&b, "ts")
	b.WriteByte(':')
	writeJSONString(&b, r.Timestamp.Format(time.RFC3339Nano))

	//ANCHOR - Message
	b.WriteByte(',')
	writeJSONString(&b, "msg")
	b.WriteByte(':')
	writeJSONString(&b, r.Message)

	//ANCHOR - Fields
	if len(r.Fields) > 0 {
		b.WriteByte(',')
		keys := sortedKeys(r.Fields)
		visited := make(map[uintptr]struct{})
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSONString(&b, k)
			b.WriteByte(':')
			f.writeJSON(&b, r.Fields[k], 0, visited)
		}
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

func (f *JsonFormatter) writeJSON(b *bytes.Buffer, v any, depth int, visited map[uintptr]struct{}) {
	if depth >= f.MaxDepth {
		writeJSONString(b, "<max_depth>")
		return
	}

	if d, ok := v.(time.Duration); ok {
		writeJSONString(b, d.String())
		return
	}

	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case string:
		writeJSONString(b, x)
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int, int8, int16, int32, int64:
		b.WriteString(strconv.FormatInt(reflect.ValueOf(x).Int(), 10))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		b.WriteString(strconv.FormatUint(reflect.ValueOf(x).Uint(), 10))
	case float32, float64:
		writeJSONFloat(b, reflect.ValueOf(x).Convert(reflect.TypeOf(float64(0))).Float())
	case time.Time:
		writeJSONString(b, x.Format(time.RFC3339Nano))
	case error:
		writeJSONString(b, x.Error())
	case fmt.Stringer:
		writeJSONString(b, x.String())
	case map[string]any:
		f.writeMapStringAny(b, x, depth, visited)
	case []any:
		f.writeSliceAny(b, x, depth, visited)
	default:
		f.writeByReflect(b, x, depth, visited)
	}
}

func (f *JsonFormatter) writeByReflect(b *bytes.Buffer, v any, depth int, visited map[uintptr]struct{}) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		b.WriteString("null")
		return
	}

	if ok, release := markAndCheck(rv, visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	switch rv.Kind() {

	//ANCHOR: REFLECT-STRING
	case reflect.String:
		writeJSONString(b, rv.String())
	//ANCHOR: REFLECT-INT
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(rv.Int(), 10))
		return
	//ANCHOR: REFLECT-UINT
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return
	//ANCHOR: REFLECT-FLOAT
	case reflect.Float32, reflect.Float64:
		writeJSONFloat(b, rv.Convert(reflect.TypeOf(float64(0))).Float())
		return
	//ANCHOR: SCALARS
	case reflect.Bool:
		if rv.Bool() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	//ANCHOR: REFLECT-INTERFACE-PTR
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			b.WriteString("null")
			return
		}
		if ok, release := markAndCheck(rv, visited); !ok {
			writeJSONString(b, "<cycle>")
			return
		} else {
			defer release()
		}
		f.writeJSON(b, rv.Elem().Interface(), depth+1, visited)
	//ANCHOR: Struct
	case reflect.Struct:
		f.writeStructByReflect(b, rv, depth, visited)
	//ANCHOR: Map
	case reflect.Map:
		f.writeMap(b, rv, depth, visited)

	//ANCHOR: SLICE, ARRAYS, BYTE
	case reflect.Slice, reflect.Array:
		// NOTE: []byte / [N]byte / alias of []byte -> base64 string
		f.writeSliceOrArrayByReflect(b, rv, depth, visited)

	default:
		writeJSONString(b, fmt.Sprintf("<unsupported:%s>", rv.Kind().String()))
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedReflectMapKeys(rv reflect.Value) []string {
	keys := rv.MapKeys()
	ss := make([]string, len(keys))
	for i, k := range keys {
		ss[i] = k.String()
	}
	sort.Strings(ss)
	return ss
}

func writeJSONString(b *bytes.Buffer, s string) {
	s = addMultilinePrefix(s)
	b.WriteString(strconv.Quote(s))
}

func writeJSONFloat(b *bytes.Buffer, f float64) {
	switch {
	case math.IsNaN(f):
		writeJSONString(b, "NaN")
	case math.IsInf(f, +1):
		writeJSONString(b, "Infinity")
	case math.IsInf(f, -1):
		writeJSONString(b, "-Infinity")
	case f == 0:
		b.WriteString("0")
	default:
		b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
	}
}

func (f *JsonFormatter) writeMapStringAny(b *bytes.Buffer, m map[string]any, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(reflect.ValueOf(m), visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	b.WriteByte('{')
	if len(m) > 0 {
		keys := sortedKeys(m)
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSONString(b, k)
			b.WriteByte(':')
			f.writeJSON(b, m[k], depth+1, visited)
		}
	}
	b.WriteByte('}')
}

func (f *JsonFormatter) writeMap(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(rv, visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	if rv.Type().Key().Kind() != reflect.String {
		writeJSONString(b, "<unsupported_map_key>")
		return
	}
	keys := sortedReflectMapKeys(rv)

	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		writeJSONString(b, k)
		b.WriteByte(':')
		f.writeJSON(b, rv.MapIndex(reflect.ValueOf(k)).Interface(), depth+1, visited)
	}
	b.WriteByte('}')
}

func (f *JsonFormatter) writeSliceAny(b *bytes.Buffer, a []any, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(reflect.ValueOf(a), visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	b.WriteByte('[')
	for i := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		f.writeJSON(b, a[i], depth+1, visited)
	}
	b.WriteByte(']')
}

func (f *JsonFormatter) writeSliceOrArrayByReflect(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(rv, visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	if rv.Type().Elem().Kind() == reflect.Uint8 {
		n := rv.Len()
		bs := make([]byte, n)
		reflect.Copy(reflect.ValueOf(bs), rv)
		writeJSONString(b, base64.StdEncoding.EncodeToString(bs))
		return
	}

	n := rv.Len()
	b.WriteByte('[')
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		f.writeJSON(b, rv.Index(i).Interface(), depth+1, visited)
	}
	b.WriteByte(']')
}

func (f *JsonFormatter) writeStructByReflect(b *bytes.Buffer, rv reflect.Value, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(rv, visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	b.WriteByte('{')
	t := rv.Type()

	type fieldInfo struct {
		key string
		idx int
	}
	fields := make([]fieldInfo, 0, rv.NumField())

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
		fields = append(fields, fieldInfo{key: key, idx: i})
	}

	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })

	for i, fi := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		writeJSONString(b, fi.key)
		b.WriteByte(':')
		f.writeJSON(b, rv.Field(fi.idx).Interface(), depth+1, visited)
	}
	b.WriteByte('}')
}
