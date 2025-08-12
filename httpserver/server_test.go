package httpserver_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rdashevsky/go-pkgs/httpserver"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []httpserver.Option
		want struct {
			hasApp bool
		}
	}{
		{
			name: "default configuration",
			opts: nil,
			want: struct{ hasApp bool }{hasApp: true},
		},
		{
			name: "with custom port",
			opts: []httpserver.Option{
				httpserver.Port(":8080"),
			},
			want: struct{ hasApp bool }{hasApp: true},
		},
		{
			name: "with timeouts",
			opts: []httpserver.Option{
				httpserver.ReadTimeout(10 * time.Second),
				httpserver.WriteTimeout(10 * time.Second),
				httpserver.ShutdownTimeout(5 * time.Second),
			},
			want: struct{ hasApp bool }{hasApp: true},
		},
		{
			name: "with prefork",
			opts: []httpserver.Option{
				httpserver.Prefork(false), // Don't actually enable prefork in tests
			},
			want: struct{ hasApp bool }{hasApp: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httpserver.New(tt.opts...)
			if server == nil {
				t.Fatal("expected server to be created")
			}
			if (server.App != nil) != tt.want.hasApp {
				t.Errorf("server.App existence = %v, want %v", server.App != nil, tt.want.hasApp)
			}
		})
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	// Create server on specific test port
	server := httpserver.New(httpserver.Port("8999"))

	// Add test route
	server.App.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Start server
	server.Start()

	// Wait for server to start or error
	select {
	case err := <-server.Notify():
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("server failed to start: %v", err)
		}
	case <-time.After(2 * time.Second):
		// Server started successfully
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	if err := server.Shutdown(); err != nil {
		t.Fatalf("failed to shutdown server: %v", err)
	}
}

func TestServer_NotifyChannel(t *testing.T) {
	server := httpserver.New(httpserver.Port(":0"))

	// Verify notify channel exists
	notify := server.Notify()
	if notify == nil {
		t.Fatal("expected notify channel to exist")
	}

	// Start server
	server.Start()

	// Wait for notification or timeout
	select {
	case <-notify:
		// Channel should be closed or have error
	case <-time.After(2 * time.Second):
		// This is also OK - server started successfully
	}

	// Shutdown
	_ = server.Shutdown()
}

func TestOptions(t *testing.T) {
	t.Run("Port option", func(t *testing.T) {
		server := httpserver.New(httpserver.Port(":9999"))
		// We can't directly test the internal address field,
		// but we can verify the server was created
		if server == nil {
			t.Fatal("expected server to be created with Port option")
		}
	})

	t.Run("Prefork option", func(t *testing.T) {
		server := httpserver.New(httpserver.Prefork(false))
		if server == nil {
			t.Fatal("expected server to be created with Prefork option")
		}
	})

	t.Run("ReadTimeout option", func(t *testing.T) {
		timeout := 30 * time.Second
		server := httpserver.New(httpserver.ReadTimeout(timeout))
		if server == nil {
			t.Fatal("expected server to be created with ReadTimeout option")
		}
		// Verify through Fiber config
		if server.App.Config().ReadTimeout != timeout {
			t.Errorf("expected ReadTimeout %v, got %v", timeout, server.App.Config().ReadTimeout)
		}
	})

	t.Run("WriteTimeout option", func(t *testing.T) {
		timeout := 30 * time.Second
		server := httpserver.New(httpserver.WriteTimeout(timeout))
		if server == nil {
			t.Fatal("expected server to be created with WriteTimeout option")
		}
		// Verify through Fiber config
		if server.App.Config().WriteTimeout != timeout {
			t.Errorf("expected WriteTimeout %v, got %v", timeout, server.App.Config().WriteTimeout)
		}
	})

	t.Run("ShutdownTimeout option", func(t *testing.T) {
		server := httpserver.New(httpserver.ShutdownTimeout(10 * time.Second))
		if server == nil {
			t.Fatal("expected server to be created with ShutdownTimeout option")
		}
		// We can't directly test the shutdown timeout without actually shutting down
	})
}

func TestServer_MultipleOptions(t *testing.T) {
	server := httpserver.New(
		httpserver.Port(":8888"),
		httpserver.ReadTimeout(20*time.Second),
		httpserver.WriteTimeout(20*time.Second),
		httpserver.ShutdownTimeout(10*time.Second),
		httpserver.Prefork(false),
	)

	if server == nil {
		t.Fatal("expected server to be created with multiple options")
	}

	// Verify Fiber app configuration
	if server.App.Config().ReadTimeout != 20*time.Second {
		t.Errorf("expected ReadTimeout 20s, got %v", server.App.Config().ReadTimeout)
	}
	if server.App.Config().WriteTimeout != 20*time.Second {
		t.Errorf("expected WriteTimeout 20s, got %v", server.App.Config().WriteTimeout)
	}
	if server.App.Config().Prefork != false {
		t.Error("expected Prefork to be false")
	}
}

func TestServer_NilApp(t *testing.T) {
	// Test edge case to ensure robust error handling
	server := httpserver.New()
	if server.App == nil {
		t.Error("expected server.App to be initialized")
	}
}

func TestServer_ConfigDefaults(t *testing.T) {
	server := httpserver.New()
	config := server.App.Config()

	// Test that defaults are applied correctly
	expectedReadTimeout := 5 * time.Second
	expectedWriteTimeout := 5 * time.Second

	if config.ReadTimeout != expectedReadTimeout {
		t.Errorf("expected default ReadTimeout %v, got %v", expectedReadTimeout, config.ReadTimeout)
	}
	if config.WriteTimeout != expectedWriteTimeout {
		t.Errorf("expected default WriteTimeout %v, got %v", expectedWriteTimeout, config.WriteTimeout)
	}
	if config.Prefork != false {
		t.Error("expected default Prefork to be false")
	}
}

func TestServer_ConcurrentAccess(t *testing.T) {
	server := httpserver.New(httpserver.Port(":0"))

	// Test concurrent access to server methods
	done := make(chan bool, 2)

	go func() {
		notify := server.Notify()
		if notify == nil {
			t.Error("expected notify channel to exist")
		}
		done <- true
	}()

	go func() {
		if server.App == nil {
			t.Error("expected server.App to exist")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
