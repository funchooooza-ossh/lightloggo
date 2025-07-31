package core

type WriteProcessor interface {
	Write(formatted []byte) error
}
