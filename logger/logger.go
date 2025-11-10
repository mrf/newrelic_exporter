package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// LogLevel represents logging levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	JSONFormat bool
	Output     io.Writer
}

// init initializes the logger with default settings
func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
}

// Setup configures the logger
func Setup(config Config) {
	// Set log level
	switch config.Level {
	case DebugLevel:
		log.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		log.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		log.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		log.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set format
	if config.JSONFormat {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	if config.Output != nil {
		log.SetOutput(config.Output)
	}
}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	return log
}

// WithField creates a new log entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields creates a new log entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Print logs a message (equivalent to Info)
func Print(args ...interface{}) {
	log.Info(args...)
}

// Printf logs a formatted message (equivalent to Infof)
func Printf(format string, args ...interface{}) {
	log.Infof(format, args...)
}
