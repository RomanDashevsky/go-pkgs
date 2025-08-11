package grpcserver_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rdashevsky/go-pkgs/grpcserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func ExampleNew() {
	// Create a new gRPC server with default settings
	server := grpcserver.New()

	// Register your services
	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

	// Start the server
	server.Start()

	// Handle shutdown
	go func() {
		if err := <-server.Notify(); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	time.Sleep(100 * time.Millisecond)
	_ = server.Shutdown()
}

func ExampleNew_withPort() {
	// Create a new gRPC server on custom port
	server := grpcserver.New(grpcserver.Port("8080"))

	// Register services
	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

	// Start the server
	server.Start()

	// Your application logic here
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	_ = server.Shutdown()
}

func ExampleServer_Start() {
	server := grpcserver.New(grpcserver.Port("50051"))

	// Register your gRPC services before starting
	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())

	// Start the server
	server.Start()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Connect to the server
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Use the connection
	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}

	fmt.Printf("Server status: %s\n", resp.Status.String())

	// Shutdown
	_ = server.Shutdown()

	// Output: Server status: SERVING
}

func ExampleServer_Shutdown() {
	server := grpcserver.New(grpcserver.Port("50052"))

	// Start server
	server.Start()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Graceful shutdown
	if err := server.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	// Wait for complete shutdown
	<-server.Notify()

	fmt.Println("Server shutdown complete")
	// Output: Server shutdown complete
}
