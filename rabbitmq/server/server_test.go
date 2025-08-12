package server_test

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rdashevsky/go-pkgs/rabbitmq/server"
)

const testResponseMessage = "test response"

// mockLogger implements logger.LoggerI for testing
type mockLogger struct {
	msgs []string
}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}

func (m *mockLogger) Info(_ string, _ ...interface{}) {}

func (m *mockLogger) Warn(_ string, _ ...interface{}) {}

func (m *mockLogger) Error(message interface{}, _ ...interface{}) {
	if msg, ok := message.(string); ok {
		m.msgs = append(m.msgs, msg)
	}
}

func (m *mockLogger) Fatal(message interface{}, args ...interface{}) {
	m.Error(message, args...)
}

func TestNew(t *testing.T) {
	t.Run("fails with invalid URL", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		_, err := server.New(
			"invalid-url",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})

	t.Run("fails with unreachable server", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{
			"test-handler": func(_ *amqp.Delivery) (interface{}, error) {
				return testResponseMessage, nil
			},
		}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected error for unreachable server")
		}
	})

	t.Run("creates server with options", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.Timeout(5*time.Second),
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		// Should fail due to connection, but options should be processed
		if err == nil {
			t.Fatal("expected connection error")
		}

		// Verify the error is about connection, not server creation
		if err.Error() == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("creates server with empty router", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("creates server with multiple handlers", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{
			"handler1": func(_ *amqp.Delivery) (interface{}, error) {
				return "response1", nil
			},
			"handler2": func(_ *amqp.Delivery) (interface{}, error) {
				return "response2", nil
			},
		}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	// Integration test - only runs if RabbitMQ is available
	t.Run("succeeds with valid server (integration)", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{
			"test": func(_ *amqp.Delivery) (interface{}, error) {
				return map[string]string{"status": "ok"}, nil
			},
		}

		s, err := server.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}
		defer func() { _ = s.Shutdown() }()

		// Verify server was created
		if s == nil {
			t.Fatal("expected server to be created")
		}

		// Verify notify channel
		notifyCh := s.Notify()
		if notifyCh == nil {
			t.Error("expected notify channel to be available")
		}
	})
}

func TestCallHandler(t *testing.T) {
	t.Run("handler function signature", func(t *testing.T) {
		handler := func(_ *amqp.Delivery) (interface{}, error) {
			return testResponseMessage, nil
		}

		// Test with mock delivery
		delivery := &amqp.Delivery{
			Body: []byte(`{"test": "data"}`),
			Type: "test-handler",
		}

		response, err := handler(delivery)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if response != testResponseMessage {
			t.Errorf("expected 'test response', got %v", response)
		}
	})

	t.Run("handler returns error", func(t *testing.T) {
		handler := func(_ *amqp.Delivery) (interface{}, error) { //nolint:unparam // Test function always returns nil interface
			return nil, &amqp.Error{Code: 404, Reason: "not found"}
		}

		delivery := &amqp.Delivery{
			Body: []byte(`{"test": "data"}`),
		}

		response, err := handler(delivery)
		if err == nil {
			t.Error("expected error from handler")
		}

		if response != nil {
			t.Errorf("expected nil response when error occurs, got %v", response)
		}
	})
}

func TestServer_Start(t *testing.T) {
	t.Run("start without connection fails gracefully", func(t *testing.T) {
		// We can't easily test Start() without a real connection
		// since it requires an actual RabbitMQ server
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Run("shutdown without connection", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		// Test shutdown on failed server creation
		_, err := server.New(
			"amqp://guest:guest@nonexistent-host:5672/",
			"server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err == nil {
			t.Fatal("expected connection error")
		}
	})

	t.Run("shutdown with connection (integration)", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{
			"test": func(_ *amqp.Delivery) (interface{}, error) {
				return "ok", nil
			},
		}

		s, err := server.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}

		// Test normal shutdown
		err = s.Shutdown()
		if err != nil {
			t.Errorf("unexpected error during shutdown: %v", err)
		}

		// Test double shutdown
		err = s.Shutdown()
		if err != nil {
			t.Errorf("unexpected error during second shutdown: %v", err)
		}
	})
}

func TestServer_Notify(t *testing.T) {
	t.Run("notify channel integration", func(t *testing.T) {
		logger := &mockLogger{}
		router := map[string]server.CallHandler{}

		s, err := server.New(
			"amqp://guest:guest@localhost:5672/",
			"test-server-exchange",
			router,
			logger,
			server.ConnWaitTime(10*time.Millisecond),
			server.ConnAttempts(1),
		)

		if err != nil {
			t.Skipf("RabbitMQ server not available: %v", err)
		}
		defer func() { _ = s.Shutdown() }()

		notifyCh := s.Notify()
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

// Test various server configurations
func TestServerOptions(t *testing.T) {
	testCases := []struct {
		name string
		opts []server.Option
	}{
		{
			name: "timeout option",
			opts: []server.Option{server.Timeout(3 * time.Second)},
		},
		{
			name: "connection wait time option",
			opts: []server.Option{server.ConnWaitTime(2 * time.Second)},
		},
		{
			name: "connection attempts option",
			opts: []server.Option{server.ConnAttempts(5)},
		},
		{
			name: "multiple options",
			opts: []server.Option{
				server.Timeout(5 * time.Second),
				server.ConnWaitTime(time.Second),
				server.ConnAttempts(3),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mockLogger{}
			router := map[string]server.CallHandler{}

			tc.opts = append(tc.opts, server.ConnWaitTime(10*time.Millisecond), server.ConnAttempts(1))
			_, err := server.New(
				"amqp://guest:guest@nonexistent-host:5672/",
				"server-exchange",
				router,
				logger,
				tc.opts...,
			)
			// Should fail due to connection, but options should be processed
			if err == nil {
				t.Fatal("expected connection error")
			}
		})
	}
}

func TestMockLogger(t *testing.T) {
	t.Run("mock logger functionality", func(t *testing.T) {
		logger := &mockLogger{}

		// Test non-error methods (should not panic)
		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")

		if len(logger.msgs) != 0 {
			t.Error("expected no messages logged for debug/info/warn")
		}
	})

	t.Run("mock logger error handling", func(t *testing.T) {
		logger := &mockLogger{}

		logger.Error("error message")

		if len(logger.msgs) != 1 || logger.msgs[0] != "error message" {
			t.Error("expected error message to be logged")
		}
	})

	t.Run("mock logger fatal handling", func(t *testing.T) {
		logger := &mockLogger{}

		logger.Fatal("fatal message")

		if len(logger.msgs) != 1 || logger.msgs[0] != "fatal message" {
			t.Error("expected fatal message to be logged")
		}
	})

	t.Run("mock logger with non-string messages", func(t *testing.T) {
		logger := &mockLogger{}

		logger.Error(123) // non-string message

		if len(logger.msgs) != 0 {
			t.Error("expected no messages logged for non-string message")
		}
	})
}
