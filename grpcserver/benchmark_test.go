package grpcserver

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func findFreeBenchPort() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "50000"
	}
	defer listener.Close()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return "50000"
	}

	return port
}

func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(Port("8080"))
	}
}

func BenchmarkServer_StartStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		port := "50100"
		server := New(Port(port))
		grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

		server.Start()
		time.Sleep(10 * time.Millisecond) // Allow server to start
		_ = server.Shutdown()

		// Wait for shutdown
		select {
		case <-server.Notify():
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	// Setup server once
	port := findFreeBenchPort()
	server := New(Port(port))
	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())
	server.Start()
	defer server.Shutdown()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Setup client
	conn, err := grpc.NewClient(
		"localhost:"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		b.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				b.Errorf("health check failed: %v", err)
			}
		}
	})
}

func BenchmarkConcurrentConnections(b *testing.B) {
	// Setup server
	port := findFreeBenchPort()
	server := New(Port(port))
	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())
	server.Start()
	defer server.Shutdown()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := grpc.NewClient(
				"localhost:"+port,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				b.Errorf("failed to connect: %v", err)
				continue
			}

			client := grpc_health_v1.NewHealthClient(conn)
			_, err = client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				b.Errorf("health check failed: %v", err)
			}

			conn.Close()
		}
	})
}
