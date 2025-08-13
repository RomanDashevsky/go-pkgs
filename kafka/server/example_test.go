package server_test

import (
	"fmt"

	"github.com/goccy/go-json"
	kafka "github.com/rdashevsky/go-pkgs/kafka"
	"github.com/rdashevsky/go-pkgs/kafka/server"
	"github.com/rdashevsky/go-pkgs/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

func ExampleNew() {
	cfg := kafka.Config{
		Brokers:  []string{"localhost:9092"},
		ClientID: "my-app-server",
		GroupID:  "my-app-server-group",
	}

	// Define handlers
	router := map[string]server.CallHandler{
		"greet": func(record *kgo.Record) (interface{}, error) {
			type Request struct {
				Name string `json:"name"`
			}

			type Response struct {
				Greeting string `json:"greeting"`
			}

			var req Request
			if err := json.Unmarshal(record.Value, &req); err != nil {
				return nil, err
			}

			return Response{
				Greeting: fmt.Sprintf("Hello, %s!", req.Name),
			}, nil
		},
	}

	logger := logger.New("info")

	server, err := server.New(cfg, "request-topic", router, logger)
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}
	defer func() { _ = server.Shutdown() }()

	// Start processing requests
	server.Start()

	fmt.Println("Kafka RPC server started successfully")
	// Output: Kafka RPC server started successfully
}
