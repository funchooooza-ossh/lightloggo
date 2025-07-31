package compressor

import (
	"compress/gzip"
	"io"
	"os"
)

type GzipCompressor struct{}

func (g *GzipCompressor) Compress(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	_, err = io.Copy(gw, in)
	gw.Close()
	return err
}

func (g *GzipCompressor) Extension() string {
	return ".gz"
}
