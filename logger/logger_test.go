package logger_test

import (
	"testing"

	"github.com/rdashevsky/go-pkgs/logger"
)

// TestLoggerCreation tests basic logger creation
func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug logger", "debug"},
		{"info logger", "info"},
		{"warn logger", "warn"},
		{"error logger", "error"},
		{"invalid logger", "invalid"},
		{"empty logger", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.New(tt.level)
			if l == nil {
				t.Error("expected logger to be created, got nil")
			}
		})
	}
}

// TestLoggerInterface ensures Logger implements LoggerI
func TestLoggerInterface(t *testing.T) {
	var _ logger.LoggerI = logger.New("info")
}

// TestLoggerMethods tests that all methods can be called without panic
func TestLoggerMethods(t *testing.T) {
	l := logger.New("debug")

	// These should not panic
	l.Debug("debug message")
	l.Debug("debug with args: %d", 42)

	l.Info("info message")
	l.Info("info with args: %s", "test")

	l.Warn("warn message")
	l.Warn("warn with args: %v", true)

	l.Error("error message")
	l.Error("error with args: %f", 3.14)

	// Test with error type
	err := &testError{msg: "test error"}
	l.Error(err)
	l.Debug(err)
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
