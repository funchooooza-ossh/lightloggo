package core

type Compressor interface {
	Compress(srcPath, dstPath string) error
	Extension() string // ".gz", ".zst", ...
}
