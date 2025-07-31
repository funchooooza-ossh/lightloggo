package main

import (
	"funchooooza-ossh/loggo/core"
	"funchooooza-ossh/loggo/core/formatter"
	"funchooooza-ossh/loggo/core/writer"
	"log"
)

func main() {
	stdout := writer.NewStdoutWriter()
	comp := writer.Gz
	fwriter, err := writer.NewFileWriter("logs/app.json", 10, 2, writer.RotateDaily, &comp)
	if err != nil {
		log.Fatalf("file error: %v", err)
	}
	json := formatter.NewJsonFormatter(nil)
	text := formatter.NewTextFormatter(nil)

	stdout_route := core.NewRouteProcessor(text, stdout, core.Debug)
	file_route := core.NewRouteProcessor(json, fwriter, core.Debug)
	logger := core.NewLogger(stdout_route, file_route)

	defer logger.Close() // вот где мы делаем закрытие очередей

	logger.Info("hello", map[string]interface{}{
		"env":   "dev",
		"stage": "test",
	})

	for i := 0; i < 1_000_000; i++ {
		logger.Info("ping", map[string]interface{}{
			"env":   "dev",
			"stage": "test",
		})
	}
}
