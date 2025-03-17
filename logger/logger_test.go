package logger

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetLevel(t *testing.T) {
	Convey("Given a default logger", t, func() {
		originalLevel := DefaultLogger.GetLevel()

		Convey("When setting the log level to debug", func() {
			SetLevel(log.DebugLevel)

			Convey("Then the logger level should be debug", func() {
				So(DefaultLogger.GetLevel(), ShouldEqual, log.DebugLevel)
			})
		})

		// Restore original level after test
		DefaultLogger.SetLevel(originalLevel)
	})
}

func TestDebug(t *testing.T) {
	Convey("Given a logger with a buffer output", t, func() {
		// Create a buffer to capture log output
		var buf bytes.Buffer
		originalLogger := DefaultLogger
		DefaultLogger = log.NewWithOptions(&buf, log.Options{
			Level:           log.DebugLevel,
			ReportCaller:    false,
			ReportTimestamp: false,
		})

		Convey("When logging a debug message", func() {
			Debug("test debug message")

			Convey("Then the message should be in the log output", func() {
				So(buf.String(), ShouldContainSubstring, "test debug message")
				So(buf.String(), ShouldContainSubstring, "debug")
			})
		})

		// Restore original logger after test
		DefaultLogger = originalLogger
	})
}

func TestInfo(t *testing.T) {
	Convey("Given a logger with a buffer output", t, func() {
		// Create a buffer to capture log output
		var buf bytes.Buffer
		originalLogger := DefaultLogger
		DefaultLogger = log.NewWithOptions(&buf, log.Options{
			Level:           log.InfoLevel,
			ReportCaller:    false,
			ReportTimestamp: false,
		})

		Convey("When logging an info message", func() {
			Info("test info message")

			Convey("Then the message should be in the log output", func() {
				So(buf.String(), ShouldContainSubstring, "test info message")
				So(buf.String(), ShouldContainSubstring, "info")
			})
		})

		// Restore original logger after test
		DefaultLogger = originalLogger
	})
}

func TestWarn(t *testing.T) {
	Convey("Given a logger with a buffer output", t, func() {
		// Create a buffer to capture log output
		var buf bytes.Buffer
		originalLogger := DefaultLogger
		DefaultLogger = log.NewWithOptions(&buf, log.Options{
			Level:           log.WarnLevel,
			ReportCaller:    false,
			ReportTimestamp: false,
		})

		Convey("When logging a warning message", func() {
			Warn("test warning message")

			Convey("Then the message should be in the log output", func() {
				So(buf.String(), ShouldContainSubstring, "test warning message")
				So(buf.String(), ShouldContainSubstring, "warn")
			})
		})

		// Restore original logger after test
		DefaultLogger = originalLogger
	})
}

func TestError(t *testing.T) {
	Convey("Given a logger with a buffer output", t, func() {
		// Create a buffer to capture log output
		var buf bytes.Buffer
		originalLogger := DefaultLogger
		DefaultLogger = log.NewWithOptions(&buf, log.Options{
			Level:           log.ErrorLevel,
			ReportCaller:    false,
			ReportTimestamp: false,
		})

		Convey("When logging an error message", func() {
			Error("test error message")

			Convey("Then the message should be in the log output", func() {
				So(buf.String(), ShouldContainSubstring, "test error message")
				So(buf.String(), ShouldContainSubstring, "error")
			})
		})

		// Restore original logger after test
		DefaultLogger = originalLogger
	})
}

func TestWithFields(t *testing.T) {
	Convey("Given a logger", t, func() {
		Convey("When creating a logger with fields", func() {
			fields := map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			}

			logger := WithFields(fields)

			Convey("Then the logger should have the fields attached", func() {
				// We can't directly access the fields in the logger
				// So we'll test by capturing output
				var buf bytes.Buffer
				logger.SetOutput(&buf)
				logger.SetLevel(log.InfoLevel)

				logger.Info("test message")

				// The actual format is "map[key1:value1 key2:42]" not "key1=value1"
				So(buf.String(), ShouldContainSubstring, "map[key1:value1")
				So(buf.String(), ShouldContainSubstring, "key2:42")
			})
		})
	})
}

func TestWithComponent(t *testing.T) {
	Convey("Given a logger", t, func() {
		Convey("When creating a logger with a component", func() {
			componentName := "test-component"
			logger := WithComponent(componentName)

			Convey("Then the logger should have the component field attached", func() {
				// We can't directly access the fields in the logger
				// So we'll test by capturing output
				var buf bytes.Buffer
				logger.SetOutput(&buf)
				logger.SetLevel(log.InfoLevel)

				logger.Info("test message")

				So(buf.String(), ShouldContainSubstring, "component=test-component")
			})
		})
	})
}
