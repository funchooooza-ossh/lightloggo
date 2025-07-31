package core

// WriteProcessor выполняет запись отформатированных логов (например, в stdout, файл или сеть).
type WriteProcessor interface {
	Write(formatted []byte) error
}

// Flushable добавляет возможность сброса буфера.
type Flushable interface {
	Flush() error
}
