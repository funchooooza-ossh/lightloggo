package main

import (
	"funchooooza-ossh/loggo/core"
	"funchooooza-ossh/loggo/core/formatter"
	"funchooooza-ossh/loggo/core/writer"
)

func main() {
	logger := core.Logger{
		Routes: []core.RouteProcessor{
			{
				Formatter:      formatter.NewJsonFormatter(),
				Writer:         writer.NewStdoutWriter(),
				LevelThreshold: core.Debug,
			},
		},
	}

	logger.Info("user_login", map[string]interface{}{
		"user_id": 123,
		"ip":      "127.0.0.1",
	})
}
