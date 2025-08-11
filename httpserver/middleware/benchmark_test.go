package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver/middleware"
)

// benchmarkLogger is a minimal logger for benchmarking
type benchmarkLogger struct {
	mu   sync.Mutex
	logs []string
}

func (b *benchmarkLogger) Debug(message interface{}, args ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, "DEBUG")
}

func (b *benchmarkLogger) Info(message string, args ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, "INFO")
}

func (b *benchmarkLogger) Warn(message string, args ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, "WARN")
}

func (b *benchmarkLogger) Error(message interface{}, args ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, "ERROR")
}

func (b *benchmarkLogger) Fatal(message interface{}, args ...interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, "FATAL")
}

// noOpLogger doesn't store logs to minimize overhead
type noOpLogger struct{}

func (n *noOpLogger) Debug(message interface{}, args ...interface{}) {}
func (n *noOpLogger) Info(message string, args ...interface{})       {}
func (n *noOpLogger) Warn(message string, args ...interface{})       {}
func (n *noOpLogger) Error(message interface{}, args ...interface{}) {}
func (n *noOpLogger) Fatal(message interface{}, args ...interface{}) {}

// BenchmarkLogger benchmarks the logger middleware with basic logging
func BenchmarkLogger(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("benchmark response")
	})

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLoggerWithStorage benchmarks logger with log storage
func BenchmarkLoggerWithStorage(b *testing.B) {
	logger := &benchmarkLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("benchmark response")
	})

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		resp.Body.Close()
	}
}

// BenchmarkLogger_GET benchmarks GET requests
func BenchmarkLogger_GET(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Get("/api/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"users": []string{"user1", "user2"}})
	})

	req := httptest.NewRequest("GET", "/api/users", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_POST benchmarks POST requests
func BenchmarkLogger_POST(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Post("/api/users", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"id": "123", "status": "created"})
	})

	body := `{"name": "John Doe", "email": "john@example.com"}`
	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_LargeResponse benchmarks responses with large payloads
func BenchmarkLogger_LargeResponse(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	largeResponse := strings.Repeat("x", 100000) // 100KB response
	app.Get("/large", func(c *fiber.Ctx) error {
		return c.SendString(largeResponse)
	})

	req := httptest.NewRequest("GET", "/large", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		resp.Body.Close()
	}
}

// BenchmarkLogger_MultipleRoutes benchmarks multiple different routes
func BenchmarkLogger_MultipleRoutes(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("home")
	})
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Post("/api/data", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"received": true})
	})
	app.Get("/static/file.txt", func(c *fiber.Ctx) error {
		return c.SendString("static content")
	})

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/", ""},
		{"GET", "/api/health", ""},
		{"POST", "/api/data", `{"test": "data"}`},
		{"GET", "/static/file.txt", ""},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			route := routes[b.N%len(routes)]
			var req *http.Request
			if route.body != "" {
				req = httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(route.method, route.path, nil)
			}

			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_WithHeaders benchmarks requests with various headers
func BenchmarkLogger_WithHeaders(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; benchmark)")
	req.Header.Set("Accept", "application/json,text/html")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	req.Header.Set("X-Request-ID", "benchmark-request-123")
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_ErrorHandling benchmarks error scenarios
func BenchmarkLogger_ErrorHandling(b *testing.B) {
	logger := &noOpLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	// Route that always returns 500
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.ErrInternalServerError
	})

	// Route that always returns 404
	app.Get("/notfound", func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	errorReq := httptest.NewRequest("GET", "/error", nil)
	notFoundReq := httptest.NewRequest("GET", "/notfound", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var req *http.Request
			if b.N%2 == 0 {
				req = errorReq
			} else {
				req = notFoundReq
			}

			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_Concurrent benchmarks concurrent requests
func BenchmarkLogger_Concurrent(b *testing.B) {
	logger := &benchmarkLogger{} // Use storage logger to test concurrency
	app := fiber.New()
	app.Use(middleware.Logger(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("concurrent test")
	})

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.SetParallelism(10) // High concurrency
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}

// BenchmarkLogger_WithoutMiddleware benchmarks baseline performance without logger
func BenchmarkLogger_WithoutMiddleware(b *testing.B) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("no logger")
	})

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := app.Test(req)
			resp.Body.Close()
		}
	})
}
