package core

type RouteProcessor struct {
	Formatter      FormatProcessor
	Writer         WriteProcessor
	LevelThreshold LogLevel
}

func (r *RouteProcessor) ShouldLog(record LogRecord) bool {
	return record.Level >= r.LevelThreshold
}

func (r *RouteProcessor) Process(record LogRecord) error {
	if !r.ShouldLog(record) {
		return nil
	}

	formatted, err := r.Formatter.Format(record)
	if err != nil {
		return err
	}

	return r.Writer.Write(formatted)
}
