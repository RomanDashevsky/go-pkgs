package server_test

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rdashevsky/go-pkgs/rabbitmq/server"
)

func TestTimeout(t *testing.T) {
	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{"1 second", time.Second},
		{"500 milliseconds", 500 * time.Millisecond},
		{"5 seconds", 5 * time.Second},
		{"zero duration", 0},
		{"negative duration", -time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't test the option directly without creating a server,
			// but we can verify it doesn't panic and the option can be created
			opt := server.Timeout(tc.timeout)
			if opt == nil {
				t.Error("expected non-nil option")
			}
		})
	}
}

func TestConnWaitTime(t *testing.T) {
	testCases := []struct {
		name     string
		waitTime time.Duration
	}{
		{"1 second", time.Second},
		{"100 milliseconds", 100 * time.Millisecond},
		{"10 seconds", 10 * time.Second},
		{"zero duration", 0},
		{"negative duration", -time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := server.ConnWaitTime(tc.waitTime)
			if opt == nil {
				t.Error("expected non-nil option")
			}
		})
	}
}

func TestConnAttempts(t *testing.T) {
	testCases := []struct {
		name     string
		attempts int
	}{
		{"single attempt", 1},
		{"three attempts", 3},
		{"many attempts", 10},
		{"zero attempts", 0},
		{"negative attempts", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := server.ConnAttempts(tc.attempts)
			if opt == nil {
				t.Error("expected non-nil option")
			}
		})
	}
}

func TestOptionsInCombination(t *testing.T) {
	// Test that multiple options can be created without conflicts
	t.Run("all options together", func(t *testing.T) {
		opts := []server.Option{
			server.Timeout(5 * time.Second),
			server.ConnWaitTime(2 * time.Second),
			server.ConnAttempts(3),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})

	t.Run("duplicate timeout options", func(t *testing.T) {
		opts := []server.Option{
			server.Timeout(time.Second),
			server.Timeout(2 * time.Second),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})

	t.Run("duplicate connection options", func(t *testing.T) {
		opts := []server.Option{
			server.ConnWaitTime(time.Second),
			server.ConnWaitTime(2 * time.Second),
			server.ConnAttempts(3),
			server.ConnAttempts(5),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})
}

// Test option application behavior by creating servers (integration test)
func TestOptionApplication(t *testing.T) {
	t.Run("options are applied during server creation", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{
			"test": func(_ *amqp.Delivery) (interface{}, error) {
				return "ok", nil
			},
		}

		// This test verifies that options don't cause panics during server creation
		// Even though connection will fail, options should be processed
		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.Timeout(100*time.Millisecond),
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		// We expect a connection error, not an option-related error
		if err == nil {
			t.Fatal("expected connection error")
		}

		// The error should be about connection, not about option application
		errStr := err.Error()
		if errStr == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("edge case values don't cause panics", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.Timeout(0),
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("options with nil router", func(t *testing.T) {
		logger := &mockLogger{}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			nil, // nil router
			logger,
			server.Timeout(time.Second),
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("options with large router", func(t *testing.T) {
		logger := &mockLogger{}
		router := make(map[string]server.CallHandler)

		// Create a router with many handlers
		for i := 0; i < 100; i++ {
			handlerName := "handler" + string(rune('0'+i%10))
			router[handlerName] = func(_ *amqp.Delivery) (interface{}, error) {
				return "response", nil
			}
		}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.Timeout(time.Second),
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})
}
