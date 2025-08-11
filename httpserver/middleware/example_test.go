package middleware_test

import (
	"fmt"
	"log"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver/middleware"
)

// exampleLogger is a simple logger implementation for examples
type exampleLogger struct{}

func (e *exampleLogger) Debug(message interface{}, args ...interface{}) {
	fmt.Printf("DEBUG: %v\n", message)
}

func (e *exampleLogger) Info(message string, args ...interface{}) {
	fmt.Printf("INFO: %s\n", message)
}

func (e *exampleLogger) Warn(message string, args ...interface{}) {
	fmt.Printf("WARN: %s\n", message)
}

func (e *exampleLogger) Error(message interface{}, args ...interface{}) {
	fmt.Printf("ERROR: %v\n", message)
}

func (e *exampleLogger) Fatal(message interface{}, args ...interface{}) {
	fmt.Printf("FATAL: %v\n", message)
}

// ExampleLogger demonstrates basic usage of the logger middleware
func ExampleLogger() {
	// Create a logger instance
	logger := &exampleLogger{}

	// Create Fiber app
	app := fiber.New()

	// Add logger middleware
	app.Use(middleware.Logger(logger))

	// Add a simple route
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Make a test request
	req := httptest.NewRequest("GET", "/hello", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	// Output:
	// INFO: 0.0.0.0 - GET /hello - 200 13
	// Response Status: 200
}

// ExampleLogger_withAPI demonstrates logger middleware with REST API endpoints
func ExampleLogger_withAPI() {
	logger := &exampleLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	// API routes
	app.Get("/api/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"users": []fiber.Map{
				{"id": 1, "name": "John Doe"},
				{"id": 2, "name": "Jane Smith"},
			},
		})
	})

	app.Post("/api/users", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"id":      3,
			"name":    "New User",
			"message": "User created successfully",
		})
	})

	// Test GET request
	getReq := httptest.NewRequest("GET", "/api/users", nil)
	getResp, _ := app.Test(getReq)
	defer getResp.Body.Close()

	// Test POST request
	postReq := httptest.NewRequest("POST", "/api/users", nil)
	postReq.Header.Set("Content-Type", "application/json")
	postResp, _ := app.Test(postReq)
	defer postResp.Body.Close()

	fmt.Printf("GET Status: %d\n", getResp.StatusCode)
	fmt.Printf("POST Status: %d\n", postResp.StatusCode)
	// Output:
	// INFO: 0.0.0.0 - GET /api/users - 200 67
	// INFO: 0.0.0.0 - POST /api/users - 201 64
	// GET Status: 200
	// POST Status: 201
}

// ExampleLogger_errorHandling demonstrates logging of error responses
func ExampleLogger_errorHandling() {
	logger := &exampleLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	// Route that returns 404
	app.Get("/notfound", func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	// Route that returns 500
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.ErrInternalServerError
	})

	// Test 404 error
	notFoundReq := httptest.NewRequest("GET", "/notfound", nil)
	notFoundResp, _ := app.Test(notFoundReq)
	defer notFoundResp.Body.Close()

	// Test 500 error
	errorReq := httptest.NewRequest("GET", "/error", nil)
	errorResp, _ := app.Test(errorReq)
	defer errorResp.Body.Close()

	fmt.Printf("404 Status: %d\n", notFoundResp.StatusCode)
	fmt.Printf("500 Status: %d\n", errorResp.StatusCode)
	// Output:
	// INFO: 0.0.0.0 - GET /notfound - 200 0
	// INFO: 0.0.0.0 - GET /error - 200 0
	// 404 Status: 404
	// 500 Status: 500
}

// stdLogger implements the LoggerI interface using standard library logger
type stdLogger struct {
	*log.Logger
}

func (s *stdLogger) Debug(message interface{}, args ...interface{}) {
	s.Printf("DEBUG: %v", message)
}

func (s *stdLogger) Info(message string, args ...interface{}) {
	s.Printf("INFO: %s", message)
}

func (s *stdLogger) Warn(message string, args ...interface{}) {
	s.Printf("WARN: %s", message)
}

func (s *stdLogger) Error(message interface{}, args ...interface{}) {
	s.Printf("ERROR: %v", message)
}

func (s *stdLogger) Fatal(message interface{}, args ...interface{}) {
	s.Printf("FATAL: %v", message)
}

// ExampleLogger_withCustomLogger demonstrates using a custom logger implementation
func ExampleLogger_withCustomLogger() {

	// Create custom logger
	customLogger := &stdLogger{Logger: log.New(log.Writer(), "[CUSTOM] ", log.LstdFlags)}

	app := fiber.New()
	app.Use(middleware.Logger(customLogger))

	app.Get("/custom", func(c *fiber.Ctx) error {
		return c.SendString("Custom logger example")
	})

	req := httptest.NewRequest("GET", "/custom", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	fmt.Printf("Custom logger response: %d\n", resp.StatusCode)
	// Output:
	// Custom logger response: 200
}

// ExampleLogger_multipleRoutes demonstrates logging across different route patterns
func ExampleLogger_multipleRoutes() {
	logger := &exampleLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	// Different types of routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Home page")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "timestamp": "2024-01-01T00:00:00Z"})
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{"user_id": id, "name": "User " + id})
	})

	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"message": "File uploaded successfully"})
	})

	// Make test requests
	requests := []struct {
		method, path string
	}{
		{"GET", "/"},
		{"GET", "/health"},
		{"GET", "/users/123"},
		{"POST", "/upload"},
	}

	for _, r := range requests {
		req := httptest.NewRequest(r.method, r.path, nil)
		resp, _ := app.Test(req)
		resp.Body.Close()
		fmt.Printf("%s %s completed\n", r.method, r.path)
	}

	// Output:
	// INFO: 0.0.0.0 - GET / - 200 9
	// GET / completed
	// INFO: 0.0.0.0 - GET /health - 200 55
	// GET /health completed
	// INFO: 0.0.0.0 - GET /users/123 - 200 35
	// GET /users/123 completed
	// INFO: 0.0.0.0 - POST /upload - 201 40
	// POST /upload completed
}

// ExampleLogger_withQueryParameters demonstrates logging requests with query parameters
func ExampleLogger_withQueryParameters() {
	logger := &exampleLogger{}
	app := fiber.New()
	app.Use(middleware.Logger(logger))

	app.Get("/search", func(c *fiber.Ctx) error {
		query := c.Query("q")
		limit := c.Query("limit", "10")

		return c.JSON(fiber.Map{
			"query":   query,
			"limit":   limit,
			"results": []string{"result1", "result2", "result3"},
		})
	})

	// Test with query parameters
	req := httptest.NewRequest("GET", "/search?q=golang&limit=5&sort=date", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	fmt.Printf("Search request status: %d\n", resp.StatusCode)
	// Output:
	// INFO: 0.0.0.0 - GET /search?q=golang&limit=5&sort=date - 200 72
	// Search request status: 200
}
