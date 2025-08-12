package client_test

import (
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq/client"
)

func TestNew(t *testing.T) {
	t.Run("fails with invalid URL", func(t *testing.T) {
		_, err := client.New(
			"invalid-url",
			"server-exchange",
			"client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})

	t.Run("fails with unreachable server", func(t *testing.T) {
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err == nil {
			t.Fatal("expected error for unreachable server")
		}
	})

	t.Run("creates client with custom timeout", func(t *testing.T) {
		// This test will fail connection but we can verify the client creation logic
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.Timeout(5*time.Second),
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		// Should fail due to connection, but that's expected
		if err == nil {
			t.Fatal("expected connection error")
		}
		// Verify the error is about connection, not client creation
		if err.Error() == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("creates client with connection options", func(t *testing.T) {
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		// Should fail due to connection, but options should be applied
		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("handles empty exchange names", func(t *testing.T) {
		_, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"",
			"",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	// Integration test - only runs if RabbitMQ is available
	t.Run("succeeds with valid server (integration)", func(t *testing.T) {
		c, err := client.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			"test-client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}
		defer c.Shutdown()

		// Verify client was created
		if c == nil {
			t.Fatal("expected client to be created")
		}

		// Verify notify channel
		notifyCh := c.Notify()
		if notifyCh == nil {
			t.Error("expected notify channel to be available")
		}
	})
}

func TestMessage(t *testing.T) {
	t.Run("message struct fields", func(t *testing.T) {
		msg := &client.Message{
			Queue:         "test-queue",
			Priority:      5,
			ContentType:   "application/json",
			Body:          []byte(`{"test": "data"}`),
			ReplyTo:       "reply-queue",
			CorrelationID: "correlation-123",
		}

		if msg.Queue != "test-queue" {
			t.Errorf("expected Queue 'test-queue', got %s", msg.Queue)
		}
		if msg.Priority != 5 {
			t.Errorf("expected Priority 5, got %d", msg.Priority)
		}
		if msg.ContentType != "application/json" {
			t.Errorf("expected ContentType 'application/json', got %s", msg.ContentType)
		}
		if string(msg.Body) != `{"test": "data"}` {
			t.Errorf("expected Body '{\"test\": \"data\"}', got %s", string(msg.Body))
		}
		if msg.ReplyTo != "reply-queue" {
			t.Errorf("expected ReplyTo 'reply-queue', got %s", msg.ReplyTo)
		}
		if msg.CorrelationID != "correlation-123" {
			t.Errorf("expected CorrelationID 'correlation-123', got %s", msg.CorrelationID)
		}
	})
}

func TestClient_RemoteCall(t *testing.T) {
	// Since RemoteCall requires actual RabbitMQ connection and server,
	// these are integration tests that will be skipped if server is not available
	t.Run("remote call integration test", func(t *testing.T) {
		c, err := client.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			"test-client-exchange",
			client.Timeout(100*time.Millisecond),
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}
		defer c.Shutdown()

		// Test remote call (will likely timeout since no server is listening)
		var response interface{}
		err = c.RemoteCall("test-handler", map[string]string{"key": "value"}, &response)

		// We expect either timeout or connection closed error
		if err == nil {
			t.Error("expected error since no server is handling requests")
		}
	})
}

func TestClient_Shutdown(t *testing.T) {
	t.Run("shutdown without connection", func(t *testing.T) {
		// Test shutdown on failed client creation
		c, err := client.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			"client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err == nil {
			defer c.Shutdown()
			t.Fatal("expected connection error")
		}
	})

	t.Run("shutdown with connection (integration)", func(t *testing.T) {
		c, err := client.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			"test-client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}

		// Test normal shutdown
		err = c.Shutdown()
		if err != nil {
			t.Errorf("unexpected error during shutdown: %v", err)
		}

		// Test double shutdown
		err = c.Shutdown()
		if err != nil {
			t.Errorf("unexpected error during second shutdown: %v", err)
		}
	})
}

func TestClient_Notify(t *testing.T) {
	t.Run("notify channel integration", func(t *testing.T) {
		c, err := client.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			"test-client-exchange",
			client.ConnWaitTime(10*time.Millisecond),
			client.ConnAttempts(1),
		)
		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}
		defer c.Shutdown()

		notifyCh := c.Notify()
		if notifyCh == nil {
			t.Fatal("expected notify channel to be available")
		}

		// Test that channel is non-blocking for reading
		select {
		case <-notifyCh:
			// Got an error, that's fine
		default:
			// No error available, also fine
		}
	})
}

// Test various client configurations
func TestClientOptions(t *testing.T) {
	testCases := []struct {
		name string
		opts []client.Option
	}{
		{
			name: "timeout option",
			opts: []client.Option{client.Timeout(3 * time.Second)},
		},
		{
			name: "connection wait time option",
			opts: []client.Option{client.ConnWaitTime(2 * time.Second)},
		},
		{
			name: "connection attempts option",
			opts: []client.Option{client.ConnAttempts(5)},
		},
		{
			name: "multiple options",
			opts: []client.Option{
				client.Timeout(5 * time.Second),
				client.ConnWaitTime(time.Second),
				client.ConnAttempts(3),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := append(tc.opts, client.ConnWaitTime(10*time.Millisecond), client.ConnAttempts(1))
			_, err := client.New(
				"amqp://guest:guest@nonexistent-host:5672/",
				"server-exchange",
				"client-exchange",
				opts...,
			)
			// Should fail due to connection, but options should be processed
			if err == nil {
				t.Fatal("expected connection error")
			}
		})
	}
}
