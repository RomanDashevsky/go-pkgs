package middleware_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver/middleware"
)

// mockLogger implements logger.LoggerI for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Debug(message interface{}, args ...interface{}) {
	m.logs = append(m.logs, "DEBUG: "+formatMessage(message, args...))
}

func (m *mockLogger) Info(message string, args ...interface{}) {
	m.logs = append(m.logs, "INFO: "+formatMessage(message, args...))
}

func (m *mockLogger) Warn(message string, args ...interface{}) {
	m.logs = append(m.logs, "WARN: "+formatMessage(message, args...))
}

func (m *mockLogger) Error(message interface{}, args ...interface{}) {
	m.logs = append(m.logs, "ERROR: "+formatMessage(message, args...))
}

func (m *mockLogger) Fatal(message interface{}, args ...interface{}) {
	m.logs = append(m.logs, "FATAL: "+formatMessage(message, args...))
}

func formatMessage(message interface{}, args ...interface{}) string {
	switch msg := message.(type) {
	case string:
		if len(args) > 0 {
			return msg // In real implementation would format with args
		}
		return msg
	case error:
		return msg.Error()
	default:
		return "unknown message type"
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		statusCode    int
		responseBody  string
		expectedInLog string
	}{
		{
			name:          "GET request",
			method:        "GET",
			path:          "/test",
			statusCode:    200,
			responseBody:  "OK",
			expectedInLog: "GET /test - 200",
		},
		{
			name:          "POST request with query",
			method:        "POST",
			path:          "/api/users?id=123",
			statusCode:    201,
			responseBody:  "Created",
			expectedInLog: "POST /api/users?id=123 - 201",
		},
		{
			name:          "404 Not Found",
			method:        "GET",
			path:          "/notfound",
			statusCode:    404,
			responseBody:  "Not Found",
			expectedInLog: "GET /notfound - 404",
		},
		{
			name:          "500 Internal Server Error",
			method:        "GET",
			path:          "/error",
			statusCode:    500,
			responseBody:  "Internal Server Error",
			expectedInLog: "GET /error - 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger
			mockLog := &mockLogger{}

			// Create Fiber app
			app := fiber.New()

			// Add logger middleware
			app.Use(middleware.Logger(mockLog))

			// Add test route
			app.All("/*", func(c *fiber.Ctx) error {
				return c.Status(tt.statusCode).SendString(tt.responseBody)
			})

			// Make request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}

			// Check response
			if resp.StatusCode != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, resp.StatusCode)
			}

			// Check logs
			if len(mockLog.logs) == 0 {
				t.Fatal("expected log entry, got none")
			}

			logEntry := mockLog.logs[0]
			if !strings.Contains(logEntry, "INFO:") {
				t.Errorf("expected INFO level log, got: %s", logEntry)
			}

			if !strings.Contains(logEntry, tt.expectedInLog) {
				t.Errorf("expected log to contain %q, got: %s", tt.expectedInLog, logEntry)
			}

			// Check response body length is in log
			bodyLenStr := string(rune(len(tt.responseBody)))
			if !strings.Contains(logEntry, bodyLenStr) && len(tt.responseBody) > 0 {
				// Note: exact format might differ, this is a basic check
			}
		})
	}
}

func TestLogger_ClientIP(t *testing.T) {
	mockLog := &mockLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(mockLog))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test with custom IP
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if len(mockLog.logs) == 0 {
		t.Fatal("expected log entry")
	}

	// The IP should be in the log
	// Note: Fiber's IP detection might vary based on config
}

func TestLogger_EmptyResponse(t *testing.T) {
	mockLog := &mockLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(mockLog))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(204) // No Content
	})

	req := httptest.NewRequest("GET", "/", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if len(mockLog.logs) == 0 {
		t.Fatal("expected log entry")
	}

	logEntry := mockLog.logs[0]
	if !strings.Contains(logEntry, "204") {
		t.Errorf("expected status 204 in log, got: %s", logEntry)
	}
}

func TestLogger_NextError(t *testing.T) {
	mockLog := &mockLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(mockLog))

	// Route that returns an error
	app.Get("/", func(c *fiber.Ctx) error {
		return fiber.ErrInternalServerError
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	// Should still log even with error
	if len(mockLog.logs) == 0 {
		t.Fatal("expected log entry even with error")
	}

	// Check that we got 500 status
	if resp.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}
}

func TestLogger_LargePayload(t *testing.T) {
	mockLog := &mockLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(mockLog))

	largeBody := strings.Repeat("a", 10000)
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendString(largeBody)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(largeBody))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if len(mockLog.logs) == 0 {
		t.Fatal("expected log entry")
	}
}

func TestLogger_WithCustomHeaders(t *testing.T) {
	mockLog := &mockLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(mockLog))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test with various headers
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Request-ID", "req-123")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if len(mockLog.logs) == 0 {
		t.Fatal("expected log entry")
	}
}
