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
	"sync"
	"time"
)

type JsonFormatter struct {
	style    *core.FormatStyle
	MaxDepth int
}

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
			KeyColor:    "\033[36m",
			ValueColor:  "\033[37m",
			Reset:       "\033[0m",
		}
	}
	return &JsonFormatter{style: style, MaxDepth: depth}
}

type fieldInfo struct {
	key       string
	idx       int
	omitEmpty bool
}

var structFieldCache sync.Map // map[reflect.Type][]fieldInfo

// ---- bytes.Buffer ----------------------------------------------------------

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func getBuf() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}
func putBuf(b *bytes.Buffer) {
	b.Reset()
	bufPool.Put(b)
}

// ---- []string  ---------------------------------------

var strSlicePool = sync.Pool{
	New: func() any { return make([]string, 0, 16) },
}

func getKeysSlice(n int) []string {
	s := strSlicePool.Get().([]string)
	if cap(s) < n {
		return make([]string, 0, n)
	}
	return s[:0]
}
func putKeysSlice(s []string) {
	for i := range s {
		s[i] = ""
	}
	if cap(s) > 4096 {
		return
	}
	strSlicePool.Put(s[:0])
}

// ---- visited map  ---------------------------------------

var visitedPool = sync.Pool{
	New: func() any { return make(map[uintptr]struct{}, 64) },
}

func getVisited() map[uintptr]struct{} {
	return visitedPool.Get().(map[uintptr]struct{})
}
func putVisited(m map[uintptr]struct{}) {
	for k := range m {
		delete(m, k)
	}
	visitedPool.Put(m)
}

func getStructFields(t reflect.Type) []fieldInfo {
	if v, ok := structFieldCache.Load(t); ok {
		return v.([]fieldInfo)
	}

	n := t.NumField()
	fields := make([]fieldInfo, 0, n)
	for i := range n {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue // unexported
		}

		key := sf.Name
		omitEmpty := false

		if tag := sf.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				key = parts[0]
			}
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					omitEmpty = true
					break
				}
			}
		}
		if key == "" {
			continue
		}
		fields = append(fields, fieldInfo{key: key, idx: i, omitEmpty: omitEmpty})
	}

	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })
	structFieldCache.Store(t, fields)
	return fields
}

func (f *JsonFormatter) Format(r core.LogRecord) ([]byte, error) {
	b := getBuf()
	defer putBuf(b)

	visited := getVisited()
	defer putVisited(visited)

	b.WriteByte('{')

	// "level"
	writeJSONString(b, "level")
	b.WriteByte(':')
	writeJSONString(b, r.Level.String())

	// ,"ts"
	b.WriteByte(',')
	writeJSONString(b, "ts")
	b.WriteByte(':')
	writeJSONString(b, r.Timestamp.Format(time.RFC3339Nano))

	// ,"msg"
	b.WriteByte(',')
	writeJSONString(b, "msg")
	b.WriteByte(':')
	writeJSONString(b, r.Message)

	// fields
	if len(r.Fields) > 0 {
		keys := getKeysSlice(len(r.Fields))
		for k := range r.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			b.WriteByte(',')
			writeJSONString(b, k)
			b.WriteByte(':')
			f.writeJSON(b, r.Fields[k], 0, visited)
		}
		putKeysSlice(keys)
	}

	b.WriteByte('}')

	out := make([]byte, b.Len())
	copy(out, b.Bytes())
	return out, nil
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
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}

	case int, int8, int16, int32, int64:
		b.WriteString(strconv.FormatInt(reflect.ValueOf(x).Int(), 10))
	case uint, uint8, uint16, uint32, uint64, uintptr:
		b.WriteString(strconv.FormatUint(reflect.ValueOf(x).Uint(), 10))
	case float32:
		writeJSONFloat(b, float64(x))
	case float64:
		writeJSONFloat(b, x)
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

func (f *JsonFormatter) writeMapStringAny(b *bytes.Buffer, m map[string]any, depth int, visited map[uintptr]struct{}) {
	if ok, release := markAndCheck(reflect.ValueOf(m), visited); !ok {
		writeJSONString(b, "<cycle>")
		return
	} else {
		defer release()
	}

	b.WriteByte('{')
	if len(m) > 0 {
		keys := getKeysSlice(len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSONString(b, k)
			b.WriteByte(':')
			f.writeJSON(b, m[k], depth+1, visited)
		}
		putKeysSlice(keys)
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
	//ANCHOR: NUMS
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(strconv.FormatInt(rv.Int(), 10))
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return
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
	case reflect.String:
		writeJSONString(b, rv.String())

	case reflect.Interface:
		if rv.IsNil() {
			b.WriteString("null")
			return
		}
		f.writeJSON(b, rv.Elem().Interface(), depth+1, visited)

	case reflect.Ptr:
		if rv.IsNil() {
			b.WriteString("null")
			return
		}
		f.writeByReflect(b, rv.Elem().Interface(), depth+1, visited)

		//ANCHOR: Struct
	case reflect.Struct:
		b.WriteByte('{')
		t := rv.Type()

		fields := getStructFields(t)

		first := true
		for _, fi := range fields {
			if fi.omitEmpty && rv.Field(fi.idx).IsZero() {
				continue
			}
			if !first {
				b.WriteByte(',')
			}
			first = false

			writeJSONString(b, fi.key)
			b.WriteByte(':')
			f.writeJSON(b, rv.Field(fi.idx).Interface(), depth+1, visited)
		}
		b.WriteByte('}')

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

	//ANCHOR: Map
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			writeJSONString(b, "<unsupported_map_key>")
			return
		}
		keys := rv.MapKeys()
		ss := getKeysSlice(len(keys))
		for i, k := range keys {
			ss[i] = k.String()
		}
		sort.Strings(ss)

		b.WriteByte('{')
		for i, k := range ss {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSONString(b, k)
			b.WriteByte(':')
			f.writeJSON(b, rv.MapIndex(reflect.ValueOf(k)).Interface(), depth+1, visited)
		}
		b.WriteByte('}')
		putKeysSlice(ss)

	//ANCHOR: SLICE, ARRAYS, BYTE
	case reflect.Slice, reflect.Array:
		// NOTE: []byte / [N]byte / alias of []byte -> base64 string
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			n := rv.Len()
			bs := make([]byte, n)
			// скопируем в bs и для slice, и для array, и для алиасов
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

	default:
		writeJSONString(b, fmt.Sprintf("<unsupported:%s>", rv.Kind().String()))
	}
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
	default:
		b.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
	}
}
