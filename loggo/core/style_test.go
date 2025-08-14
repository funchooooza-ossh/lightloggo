package core

import (
	"reflect"
	"testing"
)

// TestNewFormatStyle provides table-driven tests for the FormatStyle constructors.
// It verifies two primary scenarios:
//  1. That the NewDefaultStyle constructor initializes an object with the correct,
//     hardcoded default values.
//  2. That the NewFormatStyle constructor correctly assigns all provided custom
//     parameters to the new object's fields.
func TestNewFormatStyle(t *testing.T) {
	// testCases defines the scenarios for our table-driven test. Each struct
	// instance represents a complete, isolated test case with a name,
	// the input to test, and the expected outcome.
	testCases := []struct {
		name     string       // name is the unique name for the sub-test.
		input    *FormatStyle // input is the actual object returned by a constructor.
		expected *FormatStyle // expected is the object we expect to get.
	}{
		// Case 1: Test the constructor for the default style.
		{
			name:  "Default Style Creation",
			input: NewDefaultStyle(),
			expected: &FormatStyle{
				ColorKeys:   false,
				ColorValues: false,
				ColorLevel:  false,
				KeyColor:    "\033[34m",
				ValueColor:  "\033[33m",
				Reset:       "\033[0m",
			},
		},
		// Case 2: Test the primary constructor with custom values.
		{
			name:  "Custom Style Creation",
			input: NewFormatStyle(false, false, true, "red", "green", "reset"),
			expected: &FormatStyle{
				ColorKeys:   false,
				ColorValues: false,
				ColorLevel:  true,
				KeyColor:    "red",
				ValueColor:  "green",
				Reset:       "reset",
			},
		},
	}

	// Iterate over all defined test cases.
	for _, tc := range testCases {
		// t.Run creates a named, isolated sub-test for each case. This improves
		// test organization and provides clearer output on failure.
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input

			// Use reflect.DeepEqual for a robust, field-by-field comparison
			// of the expected and actual structs.
			if !reflect.DeepEqual(tc.expected, actual) {
				// On failure, t.Errorf provides a detailed message showing exactly
				// what was expected and what was actually received. The %#v format
				// is used to print the structs in their Go syntax representation.
				t.Errorf("Mismatched result.\nexpected: %#v\n\ngot:      %#v", tc.expected, actual)
			}
		})
	}
}
