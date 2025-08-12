package grpcserver

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func TestNew(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		server := New()
		if server == nil {
			t.Fatal("expected server to be created")
		}
		if server.App == nil {
			t.Fatal("expected grpc server to be initialized")
		}
		if server.address != ":80" {
			t.Errorf("expected default address :80, got %s", server.address)
		}
		if server.notify == nil {
			t.Fatal("expected notify channel to be initialized")
		}
	})

	t.Run("with custom port", func(t *testing.T) {
		server := New(Port("8080"))
		if server.address != ":8080" {
			t.Errorf("expected address :8080, got %s", server.address)
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		server := New(Port("9090"))
		if server.address != ":9090" {
			t.Errorf("expected address :9090, got %s", server.address)
		}
	})
}

func TestServer_Start(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		port := findFreePort(t)
		server := New(Port(port))

		// Register health service for testing
		grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

		server.Start()

		// Wait for server to start
		time.Sleep(100 * time.Millisecond)

		// Try to connect to the server
		conn, err := grpc.NewClient(
			"localhost:"+port,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			t.Fatalf("failed to connect to server: %v", err)
		}
		defer func() { _ = conn.Close() }()

		// Check health
		healthClient := grpc_health_v1.NewHealthClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}

		if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Errorf("expected SERVING status, got %v", resp.Status)
		}

		// Shutdown the server
		err = server.Shutdown()
		if err != nil {
			t.Fatalf("failed to shutdown server: %v", err)
		}
	})

	t.Run("port already in use", func(t *testing.T) {
		port := findFreePort(t)

		// Start first server
		server1 := New(Port(port))
		server1.Start()
		defer func() { _ = server1.Shutdown() }()

		// Wait for first server to start
		time.Sleep(100 * time.Millisecond)

		// Try to start second server on same port
		server2 := New(Port(port))
		server2.Start()

		// Should receive error on notify channel
		select {
		case err := <-server2.Notify():
			if err == nil {
				t.Fatal("expected error for port already in use")
			}
			if !errors.Is(err, net.ErrClosed) && !isAddressInUseError(err) {
				t.Errorf("expected address in use error, got: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("invalid address", func(t *testing.T) {
		server := &Server{
			App:     grpc.NewServer(),
			notify:  make(chan error, 1),
			address: "invalid:address:format",
		}

		server.Start()

		select {
		case err := <-server.Notify():
			if err == nil {
				t.Fatal("expected error for invalid address")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for error")
		}
	})
}

func TestServer_Notify(t *testing.T) {
	t.Run("returns notify channel", func(t *testing.T) {
		server := New()
		ch := server.Notify()
		if ch == nil {
			t.Fatal("expected notify channel to be returned")
		}

		// Verify it's the same channel
		if ch != server.notify {
			t.Error("expected same channel to be returned")
		}
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		port := findFreePort(t)
		server := New(Port(port))

		// Register health service
		grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

		server.Start()

		// Wait for server to start
		time.Sleep(100 * time.Millisecond)

		// Create a client connection
		conn, err := grpc.NewClient(
			"localhost:"+port,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			t.Fatalf("failed to connect: %v", err)
		}
		defer func() { _ = conn.Close() }()

		// Shutdown the server
		err = server.Shutdown()
		if err != nil {
			t.Fatalf("shutdown failed: %v", err)
		}

		// Try to make a request after shutdown
		healthClient := grpc_health_v1.NewHealthClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err = healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err == nil {
			t.Fatal("expected error after shutdown")
		}

		// Check that the error is related to connection/shutdown
		st, ok := status.FromError(err)
		if !ok || (st.Code() != codes.Unavailable && st.Code() != codes.Canceled && st.Code() != codes.DeadlineExceeded) {
			t.Errorf("expected Unavailable, Canceled or DeadlineExceeded error, got: %v", err)
		}
	})

	t.Run("shutdown without start", func(t *testing.T) {
		server := New()
		err := server.Shutdown()
		if err != nil {
			t.Fatalf("unexpected error on shutdown without start: %v", err)
		}
	})

	t.Run("multiple shutdowns", func(t *testing.T) {
		port := findFreePort(t)
		server := New(Port(port))

		server.Start()
		time.Sleep(100 * time.Millisecond)

		// First shutdown
		err := server.Shutdown()
		if err != nil {
			t.Fatalf("first shutdown failed: %v", err)
		}

		// Second shutdown should not panic
		err = server.Shutdown()
		if err != nil {
			t.Fatalf("second shutdown failed: %v", err)
		}
	})
}

func TestIntegration(t *testing.T) {
	t.Run("full lifecycle", func(t *testing.T) {
		port := findFreePort(t)
		server := New(Port(port))

		// Register health service
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(server.App, healthServer)

		// Start server
		server.Start()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Connect to server
		conn, err := grpc.NewClient(
			"localhost:"+port,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			t.Fatalf("failed to connect: %v", err)
		}
		defer func() { _ = conn.Close() }()

		// Check health
		healthClient := grpc_health_v1.NewHealthClient(conn)
		ctx := context.Background()

		resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
			Service: "",
		})
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}

		if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Errorf("expected SERVING status, got %v", resp.Status)
		}

		// Set service to NOT_SERVING
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		resp, err = healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}

		if resp.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
			t.Errorf("expected NOT_SERVING status, got %v", resp.Status)
		}

		// Shutdown
		err = server.Shutdown()
		if err != nil {
			t.Fatalf("shutdown failed: %v", err)
		}

		// Wait for shutdown to complete
		select {
		case <-server.Notify():
			// Server stopped
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for server to stop")
		}
	})
}

// Helper functions

func findFreePort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", ":0") //nolint:gosec // G102: Test code needs to bind to all interfaces
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to split host and port: %v", err)
	}

	return port
}

func isAddressInUseError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common "address already in use" error messages
	errStr := err.Error()
	return strings.Contains(errStr, "address already in use") ||
		strings.Contains(errStr, "bind: address already in use")
}
