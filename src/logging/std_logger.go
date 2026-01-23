package logging

import (
	"github.com/senorprogrammer/til/src"
)

// StdLogger implements Logger using the existing src logging functions
type StdLogger struct{}

// NewStdLogger creates a new StdLogger
func NewStdLogger() *StdLogger {
	return &StdLogger{}
}

func (l *StdLogger) Info(msg string) {
	src.Info(msg)
}

func (l *StdLogger) Progress(msg string) {
	src.Progress(msg)
}

func (l *StdLogger) Defeat(err error) {
	src.Defeat(err)
}

func (l *StdLogger) Victory(msg string) {
	src.Victory(msg)
}

// LoggedError wraps an error with logging context
type LoggedError struct {
	Err    error
	Logger Logger
}

func (e *LoggedError) Error() string {
	return e.Err.Error()
}

// HandleError logs the error using the logger's Defeat method
func HandleError(err error, logger Logger) {
	if err != nil {
		logger.Defeat(err)
	}
}
