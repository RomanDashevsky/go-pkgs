// Package kafka provides a Kafka client wrapper with producer and consumer support using franz-go.
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Config holds the configuration for a Kafka connection.
// It specifies the broker URLs, retry parameters, and timing.
type Config struct {
	Brokers     []string
	Timeout     time.Duration
	RetryDelay  time.Duration
	MaxRetries  int
	ClientID    string
	GroupID     string
	AutoCommit  bool
	StartOffset int64
}

// Connection represents a Kafka connection with a client.
// It manages the Kafka client lifecycle and provides retry capabilities.
type Connection struct {
	Config
	Client *kgo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConnection creates a new Kafka connection instance with the specified configuration.
// The connection is not established until Connect is called.
//
// Parameters:
//   - cfg: connection configuration including brokers, retry parameters, and timing
//
// Example:
//
//	cfg := kafka.Config{
//		Brokers:     []string{"localhost:9092"},
//		Timeout:     30 * time.Second,
//		RetryDelay:  2 * time.Second,
//		MaxRetries:  3,
//		ClientID:    "my-app",
//		GroupID:     "my-group",
//		AutoCommit:  true,
//		StartOffset: kgo.SeekEnd(),
//	}
//	conn := kafka.NewConnection(cfg)
//	err := conn.Connect(ctx)
func NewConnection(cfg Config) *Connection {
	// Set defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "kafka-client"
	}
	if cfg.StartOffset == 0 {
		cfg.StartOffset = -1 // Use -1 for end, -2 for beginning
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Connection{
		Config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect establishes a connection to Kafka brokers.
// It will retry the connection based on the configured MaxRetries and RetryDelay.
// If all attempts fail, it returns the last error encountered.
func (c *Connection) Connect(ctx context.Context) error {
	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
		kgo.ClientID(c.ClientID),
		kgo.RequestTimeoutOverhead(c.Timeout),
	}

	if c.GroupID != "" {
		opts = append(opts, kgo.ConsumerGroup(c.GroupID))
		switch c.StartOffset {
		case -1:
			opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))
		case -2:
			opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
		default:
			opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().At(c.StartOffset)))
		}

		if c.AutoCommit {
			opts = append(opts, kgo.AutoCommitInterval(1*time.Second))
		} else {
			opts = append(opts, kgo.DisableAutoCommit())
		}
	}

	var err error
	for i := 0; i <= c.MaxRetries; i++ {
		c.Client, err = kgo.NewClient(opts...)
		if err == nil {
			// Just return on successful client creation for now
			// In practice, the client will handle connection issues
			return nil
		}

		if i < c.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.RetryDelay):
			}
		}
	}

	if err != nil {
		return fmt.Errorf("kafka - Connect - failed after %d attempts: %w", c.MaxRetries+1, err)
	}

	return nil
}

// Close gracefully closes the Kafka connection.
func (c *Connection) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.Client != nil {
		c.Client.Close()
	}
}

// Context returns the connection context
func (c *Connection) Context() context.Context {
	return c.ctx
}
