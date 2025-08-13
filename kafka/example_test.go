package kafka_test

import (
	"context"
	"fmt"
	"time"

	kafka "github.com/rdashevsky/go-pkgs/kafka"
)

func ExampleNewConnection() {
	cfg := kafka.Config{
		Brokers:     []string{"localhost:9092"},
		Timeout:     30 * time.Second,
		RetryDelay:  2 * time.Second,
		MaxRetries:  3,
		ClientID:    "my-application",
		GroupID:     "my-consumer-group",
		AutoCommit:  true,
		StartOffset: -2, // Start from beginning
	}

	conn := kafka.NewConnection(cfg)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := conn.Connect(ctx)
	if err != nil {
		fmt.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("Connected to Kafka successfully")
	// Output: Connected to Kafka successfully
}
