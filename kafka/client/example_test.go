package client_test

import (
	"context"
	"fmt"
	"time"

	kafka "github.com/rdashevsky/go-pkgs/kafka"
	"github.com/rdashevsky/go-pkgs/kafka/client"
)

func ExampleNew() {
	cfg := kafka.Config{
		Brokers:     []string{"localhost:9092"},
		Timeout:     30 * time.Second,
		RetryDelay:  2 * time.Second,
		MaxRetries:  3,
		ClientID:    "my-app-client",
		GroupID:     "my-app-client-group",
		AutoCommit:  true,
		StartOffset: -1, // Start from end
	}

	client, err := client.New(cfg, "request-topic", "reply-topic", client.CallTimeout(10*time.Second))
	if err != nil {
		fmt.Printf("Failed to create client: %v", err)
		return
	}
	defer func() { _ = client.Shutdown() }()

	// Client is ready to make RPC calls
	fmt.Println("Kafka RPC client created successfully")
	// Output: Kafka RPC client created successfully
}

func ExampleClient_RemoteCall() {
	cfg := kafka.Config{
		Brokers:  []string{"localhost:9092"},
		ClientID: "example-client",
		GroupID:  "example-group",
	}

	client, err := client.New(cfg, "requests", "replies")
	if err != nil {
		fmt.Printf("Failed to create client: %v", err)
		return
	}
	defer func() { _ = client.Shutdown() }()

	type Request struct {
		Name string `json:"name"`
	}

	type Response struct {
		Greeting string `json:"greeting"`
	}

	req := Request{Name: "World"}
	var resp Response

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.RemoteCall(ctx, "greet", req, &resp)
	if err != nil {
		fmt.Printf("RPC call failed: %v", err)
		return
	}

	fmt.Printf("Response: %s", resp.Greeting)
}
