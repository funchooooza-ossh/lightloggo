package core

type FormatProcessor interface {
	Format(record LogRecord) ([]byte, error)
}
