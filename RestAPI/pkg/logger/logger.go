package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger with the given configuration
func Init(level, format, outputPath string) error {
	log = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(logLevel)

	// Set log format
	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	if outputPath != "" && outputPath != "stdout" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		// Write to both file and stdout
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	} else {
		log.SetOutput(os.Stdout)
	}

	return nil
}

// Get returns the logger instance
func Get() *logrus.Logger {
	if log == nil {
		log = logrus.New()
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	return log
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

// WithFields creates an entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Get().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Get().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Get().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}
