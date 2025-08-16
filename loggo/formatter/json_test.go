package formatter

import (
	"bytes"
	"math"
	"testing"
)

func TestWriteJsonString(t *testing.T) {
	t.Run(
		"writeJSONString", func(t *testing.T) {
			testCases := []struct {
				name     string
				input    string
				expected string
			}{
				{
					name:     "simple string",
					input:    "hello world",
					expected: `"hello world"`,
				},
				{
					name:     "single newline string",
					input:    "hello\nworld",
					expected: `"hello\n| world"`,
				},
				{
					name:     "multiple newline string",
					input:    "hello\nall\nover\nthe\nworld",
					expected: `"hello\n| all\n| over\n| the\n| world"`,
				},
				{
					name:     "special character string",
					input:    `say "hello"`,
					expected: `"say \"hello\""`,
				},
				{
					name:     "backslash string",
					input:    `C:\Users\JohnDoe`,
					expected: `"C:\\Users\\JohnDoe"`,
				},
				{
					name:     "empty string",
					input:    "",
					expected: `""`,
				},
			}
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					var b bytes.Buffer
					writeJSONString(&b, tc.input)
					actual := b.String()

					AssertEqualString(t, tc.expected, actual)
				})
			}
		},
	)
}

func TestWriteJsonFloat(t *testing.T) {
	t.Run("writeJSONFloat", func(t *testing.T) {
		testsCases := []struct {
			name     string
			input    float64
			expected string
		}{
			{
				name:     "common float",
				input:    1.23,
				expected: `1.23`,
			},
			{
				name:     "zero",
				input:    0.0,
				expected: `0`,
			},
			{
				name:     "negative zero",
				input:    math.Copysign(0.0, -1),
				expected: `0`,
			},
			{
				name:     "integer as float",
				input:    42.0,
				expected: `42`,
			},
			{
				name:     "large number",
				input:    123456789.12345678,
				expected: `1.2345678912345678e+08`,
			},
			{
				name:     "large number rounding",
				input:    123456789.123456789,
				expected: `1.2345678912345679e+08`,
			},
			{
				name:     "small non-zero number",
				input:    0.0000000000000001,
				expected: `1e-16`,
			},
			{
				name:     "float with high precision",
				input:    3.1415926535,
				expected: `3.1415926535`,
			},
			{
				name:     "NaN float",
				input:    math.NaN(),
				expected: `"NaN"`,
			},
			{
				name:     "positive infinity float",
				input:    math.Inf(0),
				expected: `"Infinity"`,
			},
			{
				name:     "negative infinity float",
				input:    math.Inf(-1),
				expected: `"-Infinity"`,
			},
		}
		for _, tc := range testsCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				writeJSONFloat(&b, tc.input)
				actual := b.String()

				AssertEqualString(t, tc.expected, actual)
			})
		}

	})
}

func TestWriteSliceAny(t *testing.T) {
	t.Run("writeSliceAny", func(t *testing.T) {
		maxDepth := 3
		formatter := NewJsonFormatter(nil, &maxDepth)

		cyclicSlice := make([]any, 2)
		cyclicSlice[0] = 1
		cyclicSlice[1] = &cyclicSlice

		testCases := []struct {
			name     string
			input    []any
			depth    int
			expected string
		}{
			{
				name:     "empty slice",
				input:    []any{},
				depth:    0,
				expected: "[]",
			},
			{
				name:     "simple slice of primitives",
				input:    []any{"a", 1, true},
				depth:    0,
				expected: `["a",1,true]`,
			},
			{
				name:     "slice with nested map",
				input:    []any{map[string]any{"key": "value"}},
				depth:    0,
				expected: `[{"key":"value"}]`,
			},
			{
				name:     "slice with nested slice",
				input:    []any{[]int{1, 2}, 3},
				depth:    0,
				expected: `[[1,2],3]`,
			},
			{
				name:     "max depth reached",
				input:    []any{[]any{[]any{"deep"}}},
				depth:    0,
				expected: `[[["<max_depth>"]]]`,
			},
			{
				name:     "cyclic slice",
				input:    cyclicSlice,
				depth:    0,
				expected: `[1,"<cycle>"]`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				formatter.writeSliceAny(&b, tc.input, tc.depth, visited)

				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})
}

func TestWriteSliceOrArrayByReflect(t *testing.T) {
	t.Run("writeSliceOrArrayByReflect", func(t *testing.T) {
		maxDepth := 3
		f := NewJsonFormatter(nil, &maxDepth)

		cyclicSlice := make([]any, 2)
		cyclicSlice[0] = 100
		cyclicSlice[1] = &cyclicSlice

		testCases := []struct {
			name     string
			input    any
			depth    int
			expected string
		}{
			{
				name:     "empty slice",
				input:    []int{},
				depth:    0,
				expected: "[]",
			},
			{
				name:     "simple slice of integers",
				input:    []int{1, 2, 3},
				depth:    0,
				expected: "[1,2,3]",
			},
			{
				name:     "simple array of strings",
				input:    [2]string{"a", "b"},
				depth:    0,
				expected: `["a","b"]`,
			},
			{
				name:     "slice of floats",
				input:    []float64{1.23, 4.56},
				depth:    0,
				expected: "[1.23,4.56]",
			},
			{
				name:     "slice of bytes is base64 encoded",
				input:    []byte("hello world"),
				depth:    0,
				expected: `"aGVsbG8gd29ybGQ="`,
			},
			{
				name:     "array of bytes is base64 encoded",
				input:    [5]byte{'h', 'e', 'l', 'l', 'o'},
				depth:    0,
				expected: `"aGVsbG8="`,
			},
			{
				name:     "nested slice",
				input:    []any{1, []int{2, 3}, 4},
				depth:    0,
				expected: `[1,[2,3],4]`,
			},
			{
				name:     "max depth reached in nested slice",
				input:    []any{[]any{[]any{"test"}}},
				depth:    0,
				expected: `[[["<max_depth>"]]]`,
			},
			{
				name:     "cyclic slice is handled",
				input:    cyclicSlice,
				depth:    0,
				expected: `[100,"<cycle>"]`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				f.writeJSON(&b, tc.input, 0, visited)

				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})
}
