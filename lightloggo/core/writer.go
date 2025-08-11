package core

// WriteProcessor выполняет запись отформатированных логов (например, в stdout, файл или сеть).
type WriteProcessor interface {
	Write(formatted []byte) error
}

// FlushableWriter — интерфейс для writer'ов с поддержкой Flush().
type FlushableWriter interface {
	Write([]byte) error
	Flush() error
}
