package server_test

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rdashevsky/go-pkgs/rabbitmq/server"
)

func ExampleNew() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"hello": func(d *amqp.Delivery) (interface{}, error) {
			return map[string]string{"message": "Hello, World!"}, nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"my-server-exchange",
		router,
		logger,
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer s.Shutdown()

	fmt.Println("Server created successfully")
	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleNew_withOptions() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"process": func(d *amqp.Delivery) (interface{}, error) {
			return map[string]interface{}{
				"status": "processed",
				"data":   string(d.Body),
			}, nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		router,
		logger,
		server.Timeout(5*time.Second),
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server with options: %v\n", err)
		return
	}
	defer s.Shutdown()

	fmt.Println("Server with custom options created")
	// Output when RabbitMQ is not available: Failed to create server with options: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleServer_Start() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"add": func(d *amqp.Delivery) (interface{}, error) {
			// Parse request and perform addition
			return map[string]int{"result": 42}, nil
		},
		"multiply": func(d *amqp.Delivery) (interface{}, error) {
			// Parse request and perform multiplication
			return map[string]int{"result": 84}, nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"math-server",
		router,
		logger,
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer s.Shutdown()

	// Start the server
	s.Start()
	fmt.Println("Server started successfully")

	// In a real application, server would run indefinitely
	// Here we just demonstrate the start

	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleServer_Notify() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"ping": func(d *amqp.Delivery) (interface{}, error) {
			return "pong", nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		router,
		logger,
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer s.Shutdown()

	// Get error notification channel
	errorCh := s.Notify()

	// Check for errors (non-blocking)
	select {
	case err := <-errorCh:
		fmt.Printf("Server error: %v\n", err)
	default:
		fmt.Println("No errors")
	}
	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleServer_Shutdown() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"task": func(d *amqp.Delivery) (interface{}, error) {
			return map[string]string{"status": "completed"}, nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"task-server",
		router,
		logger,
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// Perform some operations...
	fmt.Println("Server working...")

	// Graceful shutdown
	err = s.Shutdown()
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	} else {
		fmt.Println("Server shutdown successfully")
	}
	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleCallHandler() {
	// Example of a simple handler
	simpleHandler := func(d *amqp.Delivery) (interface{}, error) {
		return "Hello, " + string(d.Body), nil
	}

	// Example of a complex handler
	complexHandler := func(d *amqp.Delivery) (interface{}, error) {
		return map[string]interface{}{
			"received_at": time.Now().Unix(),
			"body_size":   len(d.Body),
			"type":        d.Type,
			"status":      "processed",
		}, nil
	}

	// Mock delivery
	delivery := &amqp.Delivery{
		Body: []byte("World"),
		Type: "greeting",
	}

	// Test simple handler
	result1, _ := simpleHandler(delivery)
	fmt.Printf("Simple result: %v\n", result1)

	// Test complex handler
	result2, _ := complexHandler(delivery)
	fmt.Printf("Complex result type: %T\n", result2)

	// Output:
	// Simple result: Hello, World
	// Complex result type: map[string]interface {}
}

func ExampleTimeout() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"slow": func(d *amqp.Delivery) (interface{}, error) {
			// Simulate slow processing
			time.Sleep(100 * time.Millisecond)
			return "processed", nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		router,
		logger,
		server.Timeout(500*time.Millisecond), // Short timeout
		server.ConnWaitTime(10*time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer s.Shutdown()

	fmt.Println("Server created with 500ms timeout")
	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleConnWaitTime() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"test": func(d *amqp.Delivery) (interface{}, error) {
			return "ok", nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		router,
		logger,
		server.ConnWaitTime(10*time.Millisecond), // Fast connection for example
		server.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer s.Shutdown()

	fmt.Println("Server created with fast connection wait time")
	// Output when RabbitMQ is not available: Failed to create server: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleConnAttempts() {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"robust": func(d *amqp.Delivery) (interface{}, error) {
			return map[string]bool{"robust": true}, nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"robust-server",
		router,
		logger,
		server.ConnAttempts(1), // Try 1 time for example
		server.ConnWaitTime(10*time.Millisecond),
	)
	if err != nil {
		fmt.Printf("Failed to create server after 1 attempts: %v\n", err)
		return
	}
	defer s.Shutdown()

	fmt.Println("Robust server connected successfully")
	// Output when RabbitMQ is not available: Failed to create server after 1 attempts: rmq_rpc server - NewServer - s.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}
