package logger_test

import (
	"testing"

	"github.com/rdashevsky/go-pkgs/logger"
)

// BenchmarkLoggerCreation benchmarks logger creation
func BenchmarkLoggerCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = logger.New("info")
	}
}

// BenchmarkLoggerDebug benchmarks Debug logging
func BenchmarkLoggerDebug(b *testing.B) {
	l := logger.New("debug")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("test message")
	}
}

// BenchmarkLoggerInfo benchmarks Info logging
func BenchmarkLoggerInfo(b *testing.B) {
	l := logger.New("info")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// BenchmarkLoggerWarn benchmarks Warn logging
func BenchmarkLoggerWarn(b *testing.B) {
	l := logger.New("warn")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Warn("test message")
	}
}

// BenchmarkLoggerError benchmarks Error logging
func BenchmarkLoggerError(b *testing.B) {
	l := logger.New("error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("test message")
	}
}

// BenchmarkLoggerWithFormatting benchmarks logging with formatting
func BenchmarkLoggerWithFormatting(b *testing.B) {
	l := logger.New("info")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("iteration %d with value %s", i, "test")
	}
}
