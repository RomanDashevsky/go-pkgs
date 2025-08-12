package goose_test

import (
	"fmt"

	"github.com/rdashevsky/go-pkgs/logger"
)

// mockLogger implements logger.LoggerI for testing
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   string
	message string
}

func (m *mockLogger) Debug(message interface{}, _ ...interface{}) {
	m.logs = append(m.logs, logEntry{
		level:   "DEBUG",
		message: fmt.Sprintf("%v", message),
	})
}

func (m *mockLogger) Info(message string, args ...interface{}) {
	formatted := fmt.Sprintf(message, args...)
	m.logs = append(m.logs, logEntry{
		level:   "INFO",
		message: formatted,
	})
}

func (m *mockLogger) Warn(message string, args ...interface{}) {
	formatted := fmt.Sprintf(message, args...)
	m.logs = append(m.logs, logEntry{
		level:   "WARN",
		message: formatted,
	})
}

func (m *mockLogger) Error(message interface{}, args ...interface{}) {
	var formatted string
	if len(args) > 0 {
		formatted = fmt.Sprintf("%v", message)
		for _, arg := range args {
			formatted += fmt.Sprintf(" %v", arg)
		}
	} else {
		switch msg := message.(type) {
		case string:
			formatted = msg
		case error:
			formatted = msg.Error()
		default:
			formatted = fmt.Sprintf("%v", message)
		}
	}

	m.logs = append(m.logs, logEntry{
		level:   "ERROR",
		message: formatted,
	})
}

func (m *mockLogger) Fatal(message interface{}, args ...interface{}) {
	var formatted string
	if len(args) > 0 {
		formatted = fmt.Sprintf("%v", message)
		for _, arg := range args {
			formatted += fmt.Sprintf(" %v", arg)
		}
	} else {
		switch msg := message.(type) {
		case string:
			formatted = msg
		case error:
			formatted = msg.Error()
		default:
			formatted = fmt.Sprintf("%v", message)
		}
	}

	m.logs = append(m.logs, logEntry{
		level:   "FATAL",
		message: formatted,
	})
}

func (m *mockLogger) getLogsByLevel(level string) []string {
	var messages []string
	for _, log := range m.logs {
		if log.level == level {
			messages = append(messages, log.message)
		}
	}
	return messages
}

// benchmarkLogger is a minimal logger for benchmarking that does nothing
// to avoid logging overhead affecting benchmark results
type benchmarkLogger struct{}

func (l *benchmarkLogger) Debug(_ interface{}, _ ...interface{}) {}
func (l *benchmarkLogger) Info(_ string, _ ...interface{})       {}
func (l *benchmarkLogger) Warn(_ string, _ ...interface{})       {}
func (l *benchmarkLogger) Error(_ interface{}, _ ...interface{}) {}
func (l *benchmarkLogger) Fatal(_ interface{}, _ ...interface{}) {}

// Ensure both loggers implement logger.LoggerI
var (
	_ logger.LoggerI = (*mockLogger)(nil)
	_ logger.LoggerI = (*benchmarkLogger)(nil)
)
