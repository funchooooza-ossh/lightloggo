// Package core provides the fundamental types, interfaces, and building blocks
// for the lightloggo logging library. It defines the contracts that other
// packages, such as formatters and writers, must implement.
package core

// FormatStyle holds all configuration options related to the visual styling
// of a log record, primarily for text-based output. It defines which parts
// of a log entry are colored and which ANSI escape codes to use.
type FormatStyle struct {
	// ColorKeys, if true, applies color to the keys of structured fields.
	ColorKeys bool
	// ColorValues, if true, applies color to the values of structured fields.
	ColorValues bool
	// ColorLevel, if true, applies color to the log level name (e.g., "INFO").
	ColorLevel bool

	// KeyColor is the ANSI escape code string used to color keys.
	KeyColor string
	// ValueColor is the ANSI escape code string used to color values.
	ValueColor string
	// Reset is the ANSI escape code string used to reset all text attributes.
	Reset string
}

// NewFormatStyle is the primary constructor for creating a fully customized
// FormatStyle object. It directly assigns the provided parameters to the new
// instance.
func NewFormatStyle(
	colorKeys, colorValues, colorLevel bool,
	keyColor, valueColor, reset string,
) *FormatStyle {

	return &FormatStyle{
		ColorKeys:   colorKeys,
		ColorValues: colorValues,
		ColorLevel:  colorLevel,
		KeyColor:    keyColor,
		ValueColor:  valueColor,
		Reset:       reset,
	}
}

// NewDefaultStyle is a convenience constructor that creates a FormatStyle
// instance with a pre-configured, sensible set of default values.
func NewDefaultStyle() *FormatStyle {
	// By default, coloring is disabled for a clean, non-colored output,
	// but standard ANSI codes for blue keys and yellow values are provided.
	return NewFormatStyle(
		false, false, false, "\033[34m", "\033[33m", "\033[0m",
	)
}
