package rabbitmq_test

import (
	"fmt"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq"
)

func ExampleNew() {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: 5 * time.Second,
		Attempts: 3,
	}

	conn := rabbitmq.New("my-exchange", cfg)
	fmt.Printf("Connection created with exchange: %s\n", conn.ConsumerExchange)
	// Output: Connection created with exchange: my-exchange
}

func ExampleNew_withMinimalConfig() {
	cfg := rabbitmq.Config{
		URL: "amqp://localhost:5672/",
	}

	conn := rabbitmq.New("simple-exchange", cfg)
	fmt.Printf("Exchange: %s, URL: %s\n", conn.ConsumerExchange, conn.URL)
	// Output: Exchange: simple-exchange, URL: amqp://localhost:5672/
}

func ExampleConnection_AttemptConnect() {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: time.Second,
		Attempts: 3,
	}

	conn := rabbitmq.New("example-exchange", cfg)

	err := conn.AttemptConnect()
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	fmt.Println("Successfully connected to RabbitMQ")

	// Clean up
	if conn.Connection != nil {
		_ = conn.Connection.Close()
	}
	// Output when RabbitMQ is not available: Failed to connect: rmq_rpc - AttemptConnect - c.connect: amqp.Dial: dial tcp [::1]:5672: connect: connection refused
}

func ExampleConfig() {
	cfg := rabbitmq.Config{
		URL:      "amqp://user:password@rabbitmq.example.com:5672/my-vhost",
		WaitTime: 10 * time.Second,
		Attempts: 5,
	}

	conn := rabbitmq.New("production-exchange", cfg)

	fmt.Printf("URL: %s\n", conn.URL)
	fmt.Printf("Wait time: %v\n", conn.WaitTime)
	fmt.Printf("Max attempts: %d\n", conn.Attempts)
	// Output:
	// URL: amqp://user:password@rabbitmq.example.com:5672/my-vhost
	// Wait time: 10s
	// Max attempts: 5
}

func ExampleConnection_AttemptConnect_withRetry() {
	cfg := rabbitmq.Config{
		URL:      "amqp://localhost:5672/",
		WaitTime: 2 * time.Second,
		Attempts: 3,
	}

	conn := rabbitmq.New("retry-exchange", cfg)

	fmt.Printf("Attempting to connect with %d attempts, %v wait time\n",
		conn.Attempts, conn.WaitTime)

	err := conn.AttemptConnect()
	if err != nil {
		fmt.Printf("Connection failed after %d attempts\n", conn.Attempts)
	} else {
		fmt.Println("Connection successful")
		if conn.Connection != nil {
			_ = conn.Connection.Close()
		}
	}
	// Output: Attempting to connect with 3 attempts, 2s wait time
	// Connection failed after 3 attempts
}
