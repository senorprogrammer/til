package logging

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string)
	Progress(msg string)
	Defeat(err error)
	Victory(msg string)
}
