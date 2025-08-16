package formatter

import (
	"reflect"
	"strconv"
	"strings"
)

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

// markAndCheck detects cyclical references in complex data structures by tracking
// the memory addresses of visited objects.
//
// It is a key utility to prevent infinite loops in recursive operations like
// serialization or deep printing of arbitrary data. The function operates by
// maintaining a 'visited' map of memory addresses (as uintptr) for objects
// currently in the traversal path.
//
// Go's reflection provides two different ways to obtain an object's address,
// necessitating a split based on the value's Kind:
//
//  1. For reference-like types (Ptr, Slice, Map, Chan, Func), the reflect.Value
//     is a descriptor that already contains a pointer to the underlying data.
//     We can access this directly and safely using rv.Pointer().
//
//  2. For value types (Struct, Array), the reflect.Value holds the data directly.
//     We must get the address *of the value itself*. This is done via
//     rv.Addr().Pointer(). This operation is only safe if the value is addressable
//     (i.e., not a temporary copy), so it is guarded by an rv.CanAddr() check.
//
// Return values:
//   - ok (bool): Returns false if a cycle is detected (the address was already
//     in the map). Otherwise, it returns true.
//   - release (func()): If an address is successfully marked as visited, this returns
//     a function that removes the address from the 'visited' map. The caller MUST
//     execute this function using 'defer' to ensure the object is correctly
//     unmarked as the recursion unwinds. This is crucial for correctly traversing
//     non-cyclic graphs (e.g., diamond dependencies).
func markAndCheck(rv reflect.Value, visited map[uintptr]struct{}) (ok bool, release func()) {
	var p uintptr

	switch rv.Kind() {
	// Group 1: Types that are headers/descriptors containing a pointer.
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		p = rv.Pointer()

	// Group 2: Value types that represent the data itself.
	case reflect.Struct, reflect.Array:
		// We can only get the address if the value is not a temporary copy.
		if rv.CanAddr() {
			p = rv.Addr().Pointer()
		}
	}

	// If we could not get a pointer (e.g., for a non-addressable struct) or
	// if the pointer is nil, we cannot track it. We proceed assuming it's not a cycle.
	if p == 0 {
		return true, func() {}
	}

	// Check if this memory address is already in our current traversal path.
	if _, seen := visited[p]; seen {
		// A cycle is detected.
		return false, func() {}
	}

	// Mark the address as visited and return a release function.
	// The caller is responsible for deferring the release.
	visited[p] = struct{}{}
	return true, func() { delete(visited, p) }
}

// addMultilinePrefix вставляет префикс "│ " после каждого перевода строки.
// Пример: "a\nb" -> "a\n│ b"
func addMultilinePrefix(s string) string {
	// нормализуем CRLF -> LF, затем вставляем префикс
	if strings.IndexByte(s, '\n') == -1 && !strings.Contains(s, "\r\n") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\n| ")
}
