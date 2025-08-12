// Package rabbitmq provides a simple RabbitMQ connection wrapper with automatic reconnection support.
package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config holds the configuration for a RabbitMQ connection.
// It specifies the connection URL, retry parameters, and timing.
type Config struct {
	URL      string
	WaitTime time.Duration
	Attempts int
}

// Connection represents a RabbitMQ connection with a channel and consumer setup.
// It manages the AMQP connection lifecycle and provides automatic retry capabilities.
type Connection struct {
	ConsumerExchange string
	Config
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Delivery   <-chan amqp.Delivery
}

// New creates a new RabbitMQ connection instance with the specified exchange and configuration.
// The connection is not established until AttemptConnect is called.
//
// Parameters:
//   - consumerExchange: the name of the exchange to consume from
//   - cfg: connection configuration including URL, retry attempts, and wait time
//
// Example:
//
//	cfg := rabbitmq.Config{
//		URL:      "amqp://guest:guest@localhost:5672/",
//		WaitTime: 5 * time.Second,
//		Attempts: 3,
//	}
//	conn := rabbitmq.New("my-exchange", cfg)
//	err := conn.AttemptConnect()
func New(consumerExchange string, cfg Config) *Connection {
	conn := &Connection{
		ConsumerExchange: consumerExchange,
		Config:           cfg,
	}

	return conn
}

// AttemptConnect tries to establish a connection to RabbitMQ.
// It will retry the connection based on the configured Attempts and WaitTime.
// If all attempts fail, it returns the last error encountered.
//
// The method will:
//  1. Establish an AMQP connection
//  2. Create a channel
//  3. Declare the exchange as fanout
//  4. Create an exclusive queue
//  5. Bind the queue to the exchange
//  6. Start consuming messages
func (c *Connection) AttemptConnect() error {
	var err error
	for i := c.Attempts; i > 0; i-- {
		if err = c.connect(); err == nil {
			break
		}

		log.Printf("RabbitMQ is trying to connect, attempts left: %d", i)
		time.Sleep(c.WaitTime)
	}

	if err != nil {
		return fmt.Errorf("rmq_rpc - AttemptConnect - c.connect: %w", err)
	}

	return nil
}

func (c *Connection) connect() error {
	var err error

	c.Connection, err = amqp.Dial(c.URL)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	c.Channel, err = c.Connection.Channel()
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	err = c.Channel.ExchangeDeclare(
		c.ConsumerExchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	queue, err := c.Channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueDeclare: %w", err)
	}

	err = c.Channel.QueueBind(
		queue.Name,
		"",
		c.ConsumerExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueBind: %w", err)
	}

	c.Delivery, err = c.Channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.Consume: %w", err)
	}

	return nil
}
