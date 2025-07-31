package main

import (
	"funchooooza-ossh/loggo/core"
	"funchooooza-ossh/loggo/core/formatter"
	"funchooooza-ossh/loggo/core/writer"
)

func main() {
	stdout := writer.NewStdoutWriter()
	json := formatter.NewJsonFormatter()

	route := core.NewRouteProcessor(json, stdout, core.Debug)
	logger := core.NewLogger(route)

	defer logger.Close() // вот где мы делаем закрытие очередей

	logger.Info("hello", map[string]interface{}{
		"env": "dev",
	})
}
