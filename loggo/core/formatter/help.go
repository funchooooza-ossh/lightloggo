package formatter

import "strconv"

func toIntString(v interface{}) string {
	switch i := v.(type) {
	case int:
		return strconv.Itoa(i)
	case int32:
		return strconv.FormatInt(int64(i), 10)
	case int64:
		return strconv.FormatInt(i, 10)
	default:
		return `"invalid_int"`
	}
}

func toFloatString(v interface{}) string {
	switch f := v.(type) {
	case float32:
		return strconv.FormatFloat(float64(f), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(f, 'f', -1, 64)
	default:
		return `"invalid_float"`
	}
}
