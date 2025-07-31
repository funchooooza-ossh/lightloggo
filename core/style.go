package core

type FormatStyle struct {
	ColorKeys   bool
	ColorValues bool
	ColorLevel  bool

	KeyColor   string // ANSI
	ValueColor string
	Reset      string
}
