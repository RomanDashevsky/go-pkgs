// Package logger provides a structured logging interface based on zerolog.
// It offers configurable log levels and automatic caller information.
package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// LoggerI defines the interface for structured logging with different levels.
type LoggerI interface {
	// Debug logs a debug message with optional arguments.
	Debug(message interface{}, args ...interface{})
	// Info logs an info message with optional arguments.
	Info(message string, args ...interface{})
	// Warn logs a warning message with optional arguments.
	Warn(message string, args ...interface{})
	// Error logs an error message with optional arguments.
	Error(message interface{}, args ...interface{})
	// Fatal logs a fatal message with optional arguments and exits the program.
	Fatal(message interface{}, args ...interface{})
}

// Logger implements LoggerI interface using zerolog as the underlying logger.
type Logger struct {
	logger *zerolog.Logger
}

var _ LoggerI = (*Logger)(nil)

// New creates a new Logger instance with the specified log level.
// Supported levels: "debug", "info", "warn", "error". Defaults to "info" for unknown levels.
//
// Example:
//
//	logger := logger.New("debug")
//	logger.Info("Application started")
func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

// Debug logs a debug-level message with optional formatting arguments.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg("debug", message, args...)
}

// Info logs an info-level message with optional formatting arguments.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(message, args...)
}

// Warn logs a warning-level message with optional formatting arguments.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}

	l.msg("error", message, args...)
}

// Fatal logs a fatal-level message with optional formatting arguments.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg("fatal", message, args...)

	os.Exit(1)
}

func (l *Logger) log(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Info().Msg(message)
	} else {
		l.logger.Info().Msgf(message, args...)
	}
}

func (l *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(msg.Error(), args...)
	case string:
		l.log(msg, args...)
	default:
		l.log(fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
