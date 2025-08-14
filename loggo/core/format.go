// Package core provides the fundamental types, interfaces, and building blocks
// for the lightloggo logging library.
package core

// FormatProcessor defines the contract for any log record formatter.
//
// Its primary responsibility is to take a structured LogRecord and transform it
// into a slice of bytes suitable for output by a Writer. Any struct that
// implements this interface can be used as a formatter in a logging route,
// allowing for custom, user-defined output formats.
type FormatProcessor interface {
	// Format takes a LogRecord and serializes it into a byte slice.
	// It returns the formatted output and a non-nil error if the formatting
	// process fails for any reason. Implementations must ensure that this
	// method is safe for concurrent use.
	Format(record LogRecord) ([]byte, error)
}
