package client_test

import (
	"fmt"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq/client"
)

func ExampleNew() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"my-server-exchange",
		"my-client-exchange",
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer c.Shutdown()

	fmt.Println("Client created successfully")
	// Output when RabbitMQ is not available: Failed to create client: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleNew_withOptions() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.Timeout(5*time.Second),
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client with options: %v\n", err)
		return
	}
	defer c.Shutdown()

	fmt.Println("Client with custom options created")
	// Output when RabbitMQ is not available: Failed to create client with options: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleMessage() {
	msg := &client.Message{
		Queue:         "task-queue",
		Priority:      5,
		ContentType:   "application/json",
		Body:          []byte(`{"action": "process", "data": {"id": 123}}`),
		ReplyTo:       "response-queue",
		CorrelationID: "req-12345",
	}

	fmt.Printf("Queue: %s\n", msg.Queue)
	fmt.Printf("Priority: %d\n", msg.Priority)
	fmt.Printf("Content-Type: %s\n", msg.ContentType)
	fmt.Printf("Body: %s\n", string(msg.Body))
	fmt.Printf("Reply-To: %s\n", msg.ReplyTo)
	fmt.Printf("Correlation-ID: %s\n", msg.CorrelationID)
	// Output:
	// Queue: task-queue
	// Priority: 5
	// Content-Type: application/json
	// Body: {"action": "process", "data": {"id": 123}}
	// Reply-To: response-queue
	// Correlation-ID: req-12345
}

func ExampleClient_RemoteCall() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.Timeout(time.Second),
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer c.Shutdown()

	// Prepare request
	request := map[string]interface{}{
		"operation": "add",
		"numbers":   []int{10, 20},
	}

	// Make remote call
	var response interface{}
	err = c.RemoteCall("math-handler", request, &response)
	if err != nil {
		fmt.Printf("Remote call failed: %v\n", err)
		return
	}

	fmt.Printf("Response: %v\n", response)
	// Output when server is not available: Remote call failed: rmq_rpc client - Client - RemoteCall - timeout
}

func ExampleClient_Notify() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer c.Shutdown()

	// Get error notification channel
	errorCh := c.Notify()

	// Check for errors (non-blocking)
	select {
	case err := <-errorCh:
		fmt.Printf("Client error: %v\n", err)
	default:
		fmt.Println("No errors")
	}
	// Output when RabbitMQ is not available: Failed to create client: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp 127.0.0.1:5672: connect: connection refused
}

func ExampleClient_Shutdown() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Perform some operations...
	fmt.Println("Client working...")

	// Graceful shutdown
	err = c.Shutdown()
	if err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	} else {
		fmt.Println("Client shutdown successfully")
	}
	// Output when RabbitMQ is not available: Failed to create client: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleTimeout() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.Timeout(500*time.Millisecond), // Short timeout for quick response
		client.ConnWaitTime(10*time.Millisecond),
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer c.Shutdown()

	fmt.Println("Client created with 500ms timeout")
	// Output when RabbitMQ is not available: Failed to create client: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleConnWaitTime() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.ConnWaitTime(10*time.Millisecond), // Fast connection for example
		client.ConnAttempts(1),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer c.Shutdown()

	fmt.Println("Client created with 2s connection wait time")
	// Output when RabbitMQ is not available: Failed to create client: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleConnAttempts() {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"server-exchange",
		"client-exchange",
		client.ConnAttempts(1), // Try 1 time for example
		client.ConnWaitTime(10*time.Millisecond),
	)
	if err != nil {
		fmt.Printf("Failed to create client after 1 attempts: %v\n", err)
		return
	}
	defer c.Shutdown()

	fmt.Println("Client connected successfully")
	// Output when RabbitMQ is not available: Failed to create client after 5 attempts: rmq_rpc client - NewClient - c.conn.AttemptConnect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}
