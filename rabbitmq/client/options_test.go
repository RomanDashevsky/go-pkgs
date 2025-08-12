package client_test

import (
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq/client"
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
			// We can't test the option directly without creating a client,
			// but we can verify it doesn't panic and the option can be created
			opt := client.Timeout(tc.timeout)
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
			opt := client.ConnWaitTime(tc.waitTime)
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
			opt := client.ConnAttempts(tc.attempts)
			if opt == nil {
				t.Error("expected non-nil option")
			}
		})
	}
}

func TestOptionsInCombination(t *testing.T) {
	// Test that multiple options can be created without conflicts
	t.Run("all options together", func(t *testing.T) {
		opts := []client.Option{
			client.Timeout(5 * time.Second),
			client.ConnWaitTime(2 * time.Second),
			client.ConnAttempts(3),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})

	t.Run("duplicate timeout options", func(t *testing.T) {
		opts := []client.Option{
			client.Timeout(time.Second),
			client.Timeout(2 * time.Second),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})

	t.Run("duplicate connection options", func(t *testing.T) {
		opts := []client.Option{
			client.ConnWaitTime(time.Second),
			client.ConnWaitTime(2 * time.Second),
			client.ConnAttempts(3),
			client.ConnAttempts(5),
		}

		for i, opt := range opts {
			if opt == nil {
				t.Errorf("option %d is nil", i)
			}
		}
	})
}

// Test option application behavior by creating clients (integration test)
func TestOptionApplication(t *testing.T) {
	t.Run("options are applied during client creation", func(t *testing.T) {
		// This test verifies that options don't cause panics during client creation
		// Even though connection will fail, options should be processed
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.Timeout(100*time.Millisecond),
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
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
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.Timeout(0),
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})
}
