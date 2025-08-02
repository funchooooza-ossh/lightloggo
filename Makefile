

PHONY: build 

build:
	cd ./loggo/export && go build -buildmode=c-shared -o ../../pyloggo/ffi/loggo.so
