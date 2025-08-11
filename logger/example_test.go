package logger_test

import (
	"github.com/rdashevsky/go-pkgs/logger"
)

// ExampleNew shows how to create a new logger
func ExampleNew() {
	l := logger.New("info")
	l.Info("Application started")
	// Output will include timestamp and caller information
}

// ExampleNew_debugLevel shows how to create a debug logger
func ExampleNew_debugLevel() {
	l := logger.New("debug")
	l.Debug("Debug information")
	l.Info("Info message")
	l.Warn("Warning message")
	l.Error("Error message")
}

// ExampleLogger_Info shows how to use Info logging
func ExampleLogger_Info() {
	l := logger.New("info")

	// Simple message
	l.Info("Server started")

	// With formatting
	l.Info("Listening on port %d", 8080)

	// With multiple arguments
	l.Info("User %s logged in from %s", "john", "192.168.1.1")
}

// ExampleLogger_Error shows how to use Error logging
func ExampleLogger_Error() {
	l := logger.New("error")

	// Log string error
	l.Error("Failed to connect to database")

	// Log error object
	err := doSomething()
	if err != nil {
		l.Error(err)
	}
}

func doSomething() error {
	// Example function that might return an error
	return nil
}

// ExampleLogger_Debug shows how to use Debug logging
func ExampleLogger_Debug() {
	l := logger.New("debug")

	// Debug messages are useful during development
	l.Debug("Starting request processing")
	l.Debug("Request headers: %v", map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "MyApp/1.0",
	})
	l.Debug("Processing completed")
}
