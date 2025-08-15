package formatter

import (
	"bytes"
	"funchooooza-ossh/loggo/core"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestHelpers is a container for testing small, pure utility functions
// used by the TextFormatter.
func TestTextHelpers(t *testing.T) {

	// --- Sub-test for padLevel ---
	t.Run("padLevel", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"pads short string", "INFO", "INFO   "},
			{"does not pad exact length string", "WARNING", "WARNING"},
			{"does not pad long string", "EXCEPTION", "EXCEPTION"},
			{"handles empty string", "", "       "},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				actual := padLevel(tc.input)
				AssertEqualString(t, tc.expected, actual)
			})
		}
	})

	// --- Sub-test for colorizeKey and colorizeValue ---
	t.Run("colorization", func(t *testing.T) {
		// Define styles for testing
		styleWithColors := &core.FormatStyle{
			ColorKeys:   true,
			ColorValues: true,
			KeyColor:    "[key_color]",
			ValueColor:  "[value_color]",
			Reset:       "[reset]",
		}
		styleWithoutColors := &core.FormatStyle{
			ColorKeys:   false,
			ColorValues: false,
		}

		// Create formatters with these styles
		formatterWithColors := TextFormatter{style: styleWithColors}
		formatterWithoutColors := TextFormatter{style: styleWithoutColors}

		// Test colorizeKey
		t.Run("colorizeKey", func(t *testing.T) {
			// With colors enabled
			coloredKey := formatterWithColors.colorizeKey("my-key")
			AssertContainsString(t, coloredKey, "[key_color]")
			AssertContainsString(t, coloredKey, "[reset]")

			// With colors disabled
			plainKey := formatterWithoutColors.colorizeKey("my-key")
			AssertEqualString(t, "my-key", plainKey)
		})

		// Test colorizeValue
		t.Run("colorizeValue", func(t *testing.T) {
			// With colors enabled
			coloredValue := formatterWithColors.colorizeValue("my-value")
			AssertContainsString(t, coloredValue, "[value_color]")
			AssertContainsString(t, coloredValue, "[reset]")

			// With colors disabled
			plainValue := formatterWithoutColors.colorizeValue("my-value")
			AssertEqualString(t, "my-value", plainValue)
		})
	})
}

// TestRenderText is a comprehensive table-driven test for the internal
// renderText function, which is the core of the TextFormatter.
func TestRenderText(t *testing.T) {
	// --- Test Setup ---

	// Define structs for testing reflection-based rendering.
	type sampleStruct struct {
		PublicField  string
		privateField int // Should be ignored
	}
	type taggedStruct struct {
		FieldA string `json:"field_a"`
		FieldB int    `json:"-"` // Should be ignored
		FieldC bool   `json:"field_c,omitempty"`
		FieldD string `json:"field_d,omitempty"` // omitempty on a zero-value string
	}

	// Create formatters for testing with and without colors.
	// Using placeholders for colors makes assertions much simpler.
	styleWithColors := &core.FormatStyle{
		ColorKeys:   true,
		ColorValues: true,
		KeyColor:    "<k>",
		ValueColor:  "<v>",
		Reset:       "</>",
	}
	formatterWithColors := TextFormatter{style: styleWithColors, MaxDepth: 5}
	formatterWithoutColors := TextFormatter{style: core.NewDefaultStyle(), MaxDepth: 5}

	// --- Cyclic Structures for Edge Case Testing ---
	cyclicMap := make(map[string]any)
	cyclicMap["self"] = cyclicMap
	cyclicSlice := make([]any, 2)
	cyclicSlice[0] = 1
	cyclicSlice[1] = cyclicSlice

	// --- Test Cases Table ---
	testCases := []struct {
		name     string // Name of the test case
		input    any    // The value to pass to renderText
		expected string // The expected output string
	}{
		// --- Primitives and Basic Types ---
		{"nil value", nil, "null"},
		{"string", "hello", `"hello"`},
		{"string with newline", "hello\nworld", `"hello\n| world"`},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"integer", 42, "42"},
		{"negative integer", -123, "-123"},
		{"float", 3.14, "3.14"},
		{"time.Duration", 5 * time.Second, "5s"},

		// --- Containers ---
		{"simple map", map[string]any{"key": "value"}, `{key: "value"}`},
		{"sorted map", map[string]any{"c": 3, "a": 1, "b": 2}, `{a: 1, b: 2, c: 3}`},
		{"simple slice", []any{1, "two", true}, `[1, "two", true]`},
		{"byte slice", []byte{1, 2, 3}, "[]byte(3)"},
		{"unsupported map key", map[int]string{1: "one"}, "<unsupported_map_key>"},

		// --- Structs and Reflection ---
		{"simple struct", sampleStruct{PublicField: "public", privateField: 99}, `{PublicField: "public"}`},
		{"tagged struct", taggedStruct{FieldA: "valA", FieldB: 1, FieldC: true}, `{field_a: "valA", field_c: true}`},
		{"tagged struct with omitempty (zero)", taggedStruct{FieldA: "valA"}, `{field_a: "valA"}`},
		{"tagged struct with omitempty (non-zero)", taggedStruct{FieldA: "valA", FieldC: true}, `{field_a: "valA", field_c: true}`},

		// --- Pointers ---
		{"nil pointer", (*int)(nil), "null"},
		{"pointer to value", &sampleStruct{PublicField: "pointed"}, `{PublicField: "pointed"}`},

		// --- Edge Cases ---
		{"max depth reached", map[string]any{"level1": map[string]any{"level2": "deep"}}, `{level1: {level2: <max_depth>}}`}, // Assuming MaxDepth=2 for this test
		{"cyclic map", cyclicMap, `{self: <cycle>}`},
		{"cyclic slice", cyclicSlice, `[1, <cycle>]`},
	}

	// --- Running Tests ---

	// Sub-test: Rendering without colors
	t.Run("without colors", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Special handling for the max depth test case
				formatter := formatterWithoutColors
				if tc.name == "max depth reached" {
					formatter.MaxDepth = 2 // Temporarily set a lower depth
				}

				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				formatter.renderText(&b, tc.input, 0, visited)

				// Используем наш хелпер AssertEqualString
				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})

	// Sub-test: Rendering with colors enabled
	t.Run("with colors", func(t *testing.T) {
		// We only need to test a few representative cases to ensure colorization works
		t.Run("simple map with colors", func(t *testing.T) {
			input := map[string]any{"key": "value"}
			expected := `'{<k>key</>: <v>"value"</>}'` // Note the placeholders
			// Replace placeholders with actual color codes for final comparison
			expected = strings.NewReplacer(
				"<k>", styleWithColors.KeyColor,
				"</k>", styleWithColors.Reset,
				"<v>", styleWithColors.ValueColor,
				"</>", styleWithColors.Reset,
				"'", "", // remove single quotes used for readability
			).Replace(expected)

			var b bytes.Buffer
			visited := make(map[uintptr]struct{})
			formatterWithColors.renderText(&b, input, 0, visited)

			// И здесь тоже используем наш хелпер
			AssertEqualString(t, expected, b.String())
		})
	})
}

// TestTextFormatter_Format provides a table-driven integration test for the public
// Format method. Unlike the unit tests for renderText, this test's primary
// goal is to verify that all parts of a log message (timestamp, level,
// message, and fields) are correctly assembled into the final output string.
// It confirms the overall layout and the integration between the formatter
// and its style configuration.
func TestTextFormatter_Format(t *testing.T) {
	// --- Test Setup ---

	// A fixed timestamp is used across all test cases to ensure that the
	// output is deterministic and can be reliably compared.
	fixedTime := time.Date(2025, 8, 14, 15, 30, 0, 0, time.UTC)

	// A custom style with colors enabled is created to verify that the
	// Format method correctly applies styling to the log level.
	// Placeholders are used for color codes for easier assertion.
	styleWithColors := core.NewFormatStyle(true, true, true, "<k>", "<v>", "<r>")
	// For this test to be precise, we need to know what r.Level.Color() returns.
	// Let's assume for this test that for the Error level, it returns "<err_color>".

	// --- Test Cases Table ---

	// testCases defines all scenarios to be tested. Each struct represents
	// a complete test case, including the formatter instance to use, the
	// input record, and the expected output.
	testCases := []struct {
		name           string         // The name of the sub-test.
		formatter      *TextFormatter // The formatter instance to be tested.
		record         core.LogRecord // The input log record.
		expected       string         // The expected exact full string output.
		expectContains []string       // A slice of substrings that must be in the output. Used for partial or non-deterministic checks.
	}{
		// Case 1: A standard log message with no structured fields.
		// Verifies the basic layout, timestamp, level padding, and message placement.
		{
			name:      "Simple message without fields",
			formatter: NewTextFormatter(nil, nil), // Use a default formatter
			record: core.LogRecord{
				Timestamp: fixedTime,
				Level:     core.Info,
				Message:   "hello world",
				Fields:    nil,
			},
			expected: "[2025-08-14 15:30:00.000] INFO    → hello world",
		},
		// Case 2: A log message that includes structured fields.
		// Verifies that the field separator "|" is added and that the fields
		// themselves are rendered correctly.
		{
			name:      "Message with simple fields",
			formatter: NewTextFormatter(nil, nil),
			record: core.LogRecord{
				Timestamp: fixedTime,
				Level:     core.Warning,
				Message:   "user logged in",
				Fields:    map[string]any{"id": 123},
			},
			expected: "[2025-08-14 15:30:00.000] WARNING → user logged in | id=123",
		},
		// Case 3: A test to verify the colorization logic.
		// Since the exact color code might be complex, we check for the presence
		// of key parts rather than a full string equality.
		{
			name:      "Level is colorized when style is applied",
			formatter: NewTextFormatter(styleWithColors, nil), // Use the formatter with colors
			record: core.LogRecord{
				Timestamp: fixedTime,
				Level:     core.Error,
				Message:   "critical failure",
			},
			// We check that all the essential non-colored parts are present,
			// plus the reset code from the style.
			expectContains: []string{
				"[2025-08-14 15:30:00.000]",
				// We would also check for the specific color code here if it's predictable
				// errorColorPlaceholder,
				"ERROR",
				styleWithColors.Reset,
				"→ critical failure",
			},
		},
	}

	// --- Test Execution ---

	// Iterate over each defined test case.
	for _, tc := range testCases {
		// t.Run creates an independent, named sub-test for each case,
		// providing clear and organized test output.
		t.Run(tc.name, func(t *testing.T) {
			resultBytes, err := tc.formatter.Format(tc.record)

			// First, assert that the formatter's contract of not returning an
			// error is upheld. We use our custom helper for this.
			AssertNoError(t, err)

			actualResult := string(resultBytes)

			// If an exact match is expected, perform a full string comparison.
			if tc.expected != "" {
				AssertEqualString(t, tc.expected, actualResult)
			}

			// If partial matches are expected, iterate and check for each substring.
			// This is useful for tests where parts of the output may vary (like colors)
			// or are too complex to hardcode.
			if len(tc.expectContains) > 0 {
				for _, sub := range tc.expectContains {
					if !strings.Contains(actualResult, sub) {
						t.Errorf("Result string does not contain expected substring %q.\nFull result: %q", sub, actualResult)
					}
				}
			}
		})
	}
}

// TestRenderComplexTypes focuses on the individual helper methods that render
// complex data structures like structs, maps, and slices.
func TestRenderComplexTypes(t *testing.T) {
	// --- Test Setup ---
	// A default formatter is sufficient as we are not testing colors here,
	// but the core rendering logic.
	formatter := NewTextFormatter(nil, nil)

	// --- Sub-test for renderStruct ---
	t.Run("renderStruct", func(t *testing.T) {
		// Define structs needed for testing struct rendering logic.
		type simpleStruct struct {
			PublicField  string
			privateField int // Should be ignored as it's unexported.
		}
		type taggedStruct struct {
			FieldC string `json:"c_field"`
			FieldA int    `json:"a_field"`
			FieldB bool   `json:"-"` // Should be skipped due to "-" tag.
			FieldD string `json:"d_field,omitempty"`
		}

		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{
				name:     "simple struct with unexported field",
				input:    simpleStruct{PublicField: "hello", privateField: 123},
				expected: `{PublicField: "hello"}`,
			},
			{
				name:     "tagged struct with sorting and skipping",
				input:    taggedStruct{FieldA: 1, FieldB: true, FieldC: "valC"},
				expected: `{a_field: 1, c_field: "valC"}`,
			},
			{
				name:     "tagged struct with omitempty (zero value)",
				input:    taggedStruct{FieldA: 1, FieldC: "valC", FieldD: ""}, // FieldD is zero
				expected: `{a_field: 1, c_field: "valC"}`,
			},
			{
				name:     "tagged struct with omitempty (non-zero value)",
				input:    taggedStruct{FieldA: 1, FieldC: "valC", FieldD: "valD"},
				expected: `{a_field: 1, c_field: "valC", d_field: "valD"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				// We pass a reflect.Value to the helper function.
				formatter.renderStruct(&b, reflect.ValueOf(tc.input), 0, visited, true)
				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})

	// --- Sub-test for renderMap ---
	t.Run("renderMap", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{
				name:     "map with string keys and sorted output",
				input:    map[string]any{"z": 9, "a": 1},
				expected: `{a: 1, z: 9}`,
			},
			{
				name:     "map with unsupported key type",
				input:    map[int]string{1: "one"},
				expected: `<unsupported_map_key>`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				formatter.renderMap(&b, reflect.ValueOf(tc.input), 0, visited, true)
				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})

	// --- Sub-test for renderSlice ---
	t.Run("renderSlice", func(t *testing.T) {

		cyclicSlice := make([]any, 2)
		cyclicSlice[0] = 100
		cyclicSlice[1] = &cyclicSlice

		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{
				name:     "simple slice of mixed types",
				input:    []any{1, "two", false},
				expected: `[1, "two", false]`,
			},
			{
				name:     "special handling for byte slice",
				input:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
				expected: "[]byte(4)",
			},
			{
				name:     "empty slice",
				input:    []int{},
				expected: "[]",
			},
			{
				name:     "cyclic slice",
				input:    cyclicSlice,
				expected: "[100, <cycle>]",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var b bytes.Buffer
				visited := make(map[uintptr]struct{})
				formatter.renderSlice(&b, reflect.ValueOf(tc.input), 0, visited, true)
				AssertEqualString(t, tc.expected, b.String())
			})
		}
	})
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertEqualString(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("Mismatched result.\nexpected: %q\n\ngot:      %q", expected, actual)
	}
}
func AssertContainsString(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected string %q to contain substring %q", s, substr)
	}
}
