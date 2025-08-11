package httpserver_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver"
)

// BenchmarkNew benchmarks server creation with default options
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = httpserver.New()
	}
}

// BenchmarkNewWithOptions benchmarks server creation with multiple options
func BenchmarkNewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = httpserver.New(
			httpserver.Port(":8080"),
			httpserver.ReadTimeout(10*time.Second),
			httpserver.WriteTimeout(10*time.Second),
		)
	}
}

// BenchmarkNewWithAllOptions benchmarks server creation with all available options
func BenchmarkNewWithAllOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = httpserver.New(
			httpserver.Port(":8080"),
			httpserver.ReadTimeout(10*time.Second),
			httpserver.WriteTimeout(10*time.Second),
			httpserver.ShutdownTimeout(5*time.Second),
			httpserver.Prefork(false),
		)
	}
}

// BenchmarkServer_Start benchmarks server startup
func BenchmarkServer_Start(b *testing.B) {
	for i := 0; i < b.N; i++ {
		server := httpserver.New(httpserver.Port(":0"))

		b.StartTimer()
		server.Start()
		b.StopTimer()

		// Clean shutdown
		server.Shutdown()
	}
}

// BenchmarkServer_NotifyChannel benchmarks accessing notify channel
func BenchmarkServer_NotifyChannel(b *testing.B) {
	server := httpserver.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.Notify()
	}
}

// BenchmarkOptionApplication benchmarks individual option applications
func BenchmarkOptionApplication(b *testing.B) {
	b.Run("Port", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = httpserver.New(httpserver.Port(":8080"))
		}
	})

	b.Run("ReadTimeout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = httpserver.New(httpserver.ReadTimeout(10 * time.Second))
		}
	})

	b.Run("WriteTimeout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = httpserver.New(httpserver.WriteTimeout(10 * time.Second))
		}
	})

	b.Run("ShutdownTimeout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = httpserver.New(httpserver.ShutdownTimeout(5 * time.Second))
		}
	})

	b.Run("Prefork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = httpserver.New(httpserver.Prefork(false))
		}
	})
}

// BenchmarkServer_RouteRegistration benchmarks adding routes to server
func BenchmarkServer_RouteRegistration(b *testing.B) {
	server := httpserver.New()
	handler := func(c *fiber.Ctx) error {
		return c.SendString("OK")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.App.Get("/test", handler)
	}
}

// BenchmarkServer_MultipleRouteRegistration benchmarks adding multiple routes
func BenchmarkServer_MultipleRouteRegistration(b *testing.B) {
	handler := func(c *fiber.Ctx) error {
		return c.SendString("OK")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := httpserver.New()
		server.App.Get("/", handler)
		server.App.Post("/create", handler)
		server.App.Put("/update", handler)
		server.App.Delete("/delete", handler)
		server.App.Get("/health", handler)
	}
}

// BenchmarkServer_StartupShutdownCycle benchmarks complete server lifecycle
func BenchmarkServer_StartupShutdownCycle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		server := httpserver.New(httpserver.Port(":0"))

		// Add a simple route
		server.App.Get("/ping", func(c *fiber.Ctx) error {
			return c.SendString("pong")
		})

		// Start server
		server.Start()

		// Wait briefly for startup
		select {
		case err := <-server.Notify():
			if err != nil && err != http.ErrServerClosed {
				b.Fatalf("Server startup failed: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			// Server started successfully
		}

		// Shutdown
		server.Shutdown()
	}
}

// BenchmarkServer_ConcurrentCreation benchmarks concurrent server creation
func BenchmarkServer_ConcurrentCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = httpserver.New(
				httpserver.Port(":0"),
				httpserver.ReadTimeout(5*time.Second),
			)
		}
	})
}

// BenchmarkServer_ConcurrentNotify benchmarks concurrent access to Notify
func BenchmarkServer_ConcurrentNotify(b *testing.B) {
	server := httpserver.New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = server.Notify()
		}
	})
}

// BenchmarkServer_MemoryAllocation benchmarks memory allocation patterns
func BenchmarkServer_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		server := httpserver.New(
			httpserver.Port(":8080"),
			httpserver.ReadTimeout(10*time.Second),
			httpserver.WriteTimeout(10*time.Second),
			httpserver.ShutdownTimeout(5*time.Second),
		)

		// Add routes to trigger more allocations
		server.App.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hello World")
		})

		// Access notify to ensure channel allocation
		_ = server.Notify()
	}
}

// BenchmarkServer_HighVolumeOptionApplication benchmarks applying many options
func BenchmarkServer_HighVolumeOptionApplication(b *testing.B) {
	// Create a large slice of options to simulate heavy configuration
	options := make([]httpserver.Option, 0, 100)

	// Add many timeout variations
	for i := 0; i < 25; i++ {
		options = append(options, httpserver.ReadTimeout(time.Duration(i+1)*time.Second))
	}
	for i := 0; i < 25; i++ {
		options = append(options, httpserver.WriteTimeout(time.Duration(i+1)*time.Second))
	}
	for i := 0; i < 25; i++ {
		options = append(options, httpserver.ShutdownTimeout(time.Duration(i+1)*time.Second))
	}
	for i := 0; i < 25; i++ {
		// Alternate port configurations
		if i%2 == 0 {
			options = append(options, httpserver.Port(":8080"))
		} else {
			options = append(options, httpserver.Port(":9090"))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpserver.New(options...)
	}
}
