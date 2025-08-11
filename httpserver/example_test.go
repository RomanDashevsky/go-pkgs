package httpserver_test

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver"
)

// Example demonstrates basic HTTP server creation and usage
func Example() {
	// Create server with default configuration
	server := httpserver.New()

	// Add a simple route
	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Start server (commented out for example)
	// server.Start()

	// Wait for shutdown signal (commented out for example)
	// <-server.Notify()
}

// ExampleNew_withCustomPort demonstrates creating a server with a custom port
func ExampleNew_withCustomPort() {
	server := httpserver.New(
		httpserver.Port(":8080"),
	)

	server.App.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleNew_withTimeouts demonstrates configuring server timeouts
func ExampleNew_withTimeouts() {
	server := httpserver.New(
		httpserver.ReadTimeout(10*time.Second),
		httpserver.WriteTimeout(10*time.Second),
		httpserver.ShutdownTimeout(5*time.Second),
	)

	server.App.Get("/slow", func(c *fiber.Ctx) error {
		// Simulate slow operation
		time.Sleep(100 * time.Millisecond)
		return c.SendString("Completed")
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleNew_fullConfiguration demonstrates a server with all configuration options
func ExampleNew_fullConfiguration() {
	server := httpserver.New(
		httpserver.Port(":8080"),
		httpserver.ReadTimeout(30*time.Second),
		httpserver.WriteTimeout(30*time.Second),
		httpserver.ShutdownTimeout(10*time.Second),
		httpserver.Prefork(false), // Set to true in production for better performance
	)

	// Add middleware and routes
	server.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleServer_Start demonstrates starting a server and handling errors
func ExampleServer_Start() {
	server := httpserver.New(httpserver.Port(":8080"))

	// Add routes before starting
	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server is running!")
	})

	// Start the server
	server.Start()

	// Wait for startup errors or shutdown
	if err := <-server.Notify(); err != nil {
		log.Printf("Server error: %v", err)
	}
}

// ExampleServer_Shutdown demonstrates graceful server shutdown
func ExampleServer_Shutdown() {
	server := httpserver.New(httpserver.Port(":8080"))

	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello!")
	})

	// Start server
	server.Start()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Gracefully shutdown the server
	if err := server.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}

// ExampleServer_restAPI demonstrates creating a REST API
func ExampleServer_restAPI() {
	server := httpserver.New(
		httpserver.Port(":8080"),
		httpserver.ReadTimeout(15*time.Second),
		httpserver.WriteTimeout(15*time.Second),
	)

	// Sample data
	users := []fiber.Map{
		{"id": 1, "name": "John Doe", "email": "john@example.com"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}

	// API routes
	api := server.App.Group("/api/v1")

	api.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": users,
		})
	})

	api.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{
			"id":      id,
			"message": "User details",
		})
	})

	api.Post("/users", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"message": "User created",
		})
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleServer_withMiddleware demonstrates adding middleware to the server
func ExampleServer_withMiddleware() {
	server := httpserver.New(httpserver.Port(":8080"))

	// Add custom middleware
	server.App.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request
		log.Printf("%s %s - %v", c.Method(), c.Path(), time.Since(start))
		return err
	})

	// Add CORS-like middleware
	server.App.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Next()
	})

	// Routes
	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello with middleware!")
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleServer_notifyChannel demonstrates using the notify channel for error handling
func ExampleServer_notifyChannel() {
	server := httpserver.New(httpserver.Port(":8080"))

	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello!")
	})

	// Start server
	server.Start()

	// Handle server lifecycle
	go func() {
		err := <-server.Notify()
		if err != nil {
			log.Printf("Server stopped with error: %v", err)
		} else {
			log.Println("Server stopped gracefully")
		}
	}()

	// Simulate running for a short time
	time.Sleep(100 * time.Millisecond)
	server.Shutdown()
}

// ExampleServer_fileServer demonstrates serving static files
func ExampleServer_fileServer() {
	server := httpserver.New(httpserver.Port(":8080"))

	// Serve static files (in a real scenario, make sure the directory exists)
	server.App.Static("/static", "./public")

	// API endpoint
	server.App.Get("/api/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":     "My App",
			"version": "1.0.0",
		})
	})

	// Default route
	server.App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the file server!")
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleServer_jsonAPI demonstrates a JSON API server
func ExampleServer_jsonAPI() {
	server := httpserver.New(
		httpserver.Port(":8080"),
		httpserver.ReadTimeout(10*time.Second),
		httpserver.WriteTimeout(10*time.Second),
	)

	// JSON request/response example
	server.App.Post("/api/data", func(c *fiber.Ctx) error {
		var payload struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		return c.JSON(fiber.Map{
			"received":  payload,
			"status":    "success",
			"timestamp": time.Now(),
		})
	})

	// server.Start()
	// <-server.Notify()
}

// ExampleServer_productsionReady demonstrates a production-ready server configuration
func ExampleServer_productionReady() {
	server := httpserver.New(
		httpserver.Port(":8080"),
		httpserver.ReadTimeout(30*time.Second),
		httpserver.WriteTimeout(30*time.Second),
		httpserver.ShutdownTimeout(15*time.Second),
		httpserver.Prefork(true), // Enable for production
	)

	// Health check endpoint
	server.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"uptime": time.Now(),
		})
	})

	// Metrics endpoint
	server.App.Get("/metrics", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"requests_total": 12345,
			"memory_usage":   "45.2MB",
		})
	})

	// Main application routes
	api := server.App.Group("/api")
	api.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version": "1.0.0",
			"build":   "20240111",
		})
	})

	// server.Start()
	// log.Printf("Production server started on port 8080")
	// <-server.Notify()
}
