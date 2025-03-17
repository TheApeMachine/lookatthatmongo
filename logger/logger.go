package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var (
	/*
		DefaultLogger is the default logger instance used throughout the application.
		It is configured with standard options for level, caller reporting, and timestamps.
	*/
	DefaultLogger = log.NewWithOptions(os.Stderr, log.Options{
		Level:           log.InfoLevel,
		ReportCaller:    true,
		ReportTimestamp: true,
	})
)

/*
SetLevel sets the logging level for the default logger.
*/
func SetLevel(level log.Level) {
	DefaultLogger.SetLevel(level)
}

/*
Debug logs a debug message with optional fields.
*/
func Debug(msg string, fields ...any) {
	DefaultLogger.Debug(msg, fields...)
}

/*
Info logs an info message with optional fields.
*/
func Info(msg string, fields ...any) {
	DefaultLogger.Info(msg, fields...)
}

/*
Warn logs a warning message with optional fields.
*/
func Warn(msg string, fields ...any) {
	DefaultLogger.Warn(msg, fields...)
}

/*
Error logs an error message with optional fields.
*/
func Error(msg string, fields ...any) {
	DefaultLogger.Error(msg, fields...)
}

/*
Fatal logs a fatal message with optional fields and exits the application.
*/
func Fatal(msg string, fields ...any) {
	DefaultLogger.Fatal(msg, fields...)
}

/*
WithFields creates a new logger with the given fields attached to all log entries.
*/
func WithFields(fields map[string]interface{}) *log.Logger {
	return DefaultLogger.With(fields)
}

/*
WithComponent creates a new logger with the component field set to identify the source.
*/
func WithComponent(component string) *log.Logger {
	return DefaultLogger.With("component", component)
}
