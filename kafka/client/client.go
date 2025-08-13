// Package client provides a Kafka RPC client implementation for making remote procedure calls.
package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	kafka "github.com/rdashevsky/go-pkgs/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ErrConnectionClosed is returned when attempting to make a remote call on a closed connection.
var ErrConnectionClosed = errors.New("kafka_rpc client - Client - RemoteCall - Connection closed")

const (
	_defaultTimeout     = 30 * time.Second
	_defaultRetryDelay  = 2 * time.Second
	_defaultMaxRetries  = 3
	_defaultCallTimeout = 10 * time.Second
)

// Message represents a Kafka message with all its properties.
// It can be used to construct messages for publishing or to inspect received messages.
type Message struct {
	Topic         string
	Key           string
	Value         []byte
	Headers       map[string]string
	Partition     int32
	Offset        int64
	Timestamp     time.Time
	CorrelationID string
	ReplyTopic    string
}

type pendingCall struct {
	done   chan struct{}
	status string
	body   []byte
	err    error
}

// Client represents a Kafka RPC client for making remote procedure calls.
// It manages the connection, handles request-response correlation, and provides timeout support.
type Client struct {
	conn         *kafka.Connection
	requestTopic string
	replyTopic   string
	error        chan error
	stop         chan struct{}

	rw    sync.RWMutex
	calls map[string]*pendingCall

	callTimeout time.Duration
}

// New creates a new Kafka RPC client with the specified configuration.
// The client establishes a connection immediately and starts consuming responses.
//
// Parameters:
//   - cfg: Kafka connection configuration
//   - requestTopic: topic name where requests will be published
//   - replyTopic: topic name where responses will be received
//   - opts: optional configuration functions
//
// Returns an error if the connection cannot be established.
func New(cfg kafka.Config, requestTopic, replyTopic string, opts ...Option) (*Client, error) {
	// Ensure we have a consumer group for replies
	if cfg.GroupID == "" {
		cfg.GroupID = fmt.Sprintf("kafka-rpc-client-%s", uuid.New().String())
	}

	conn := kafka.NewConnection(cfg)

	c := &Client{
		conn:         conn,
		requestTopic: requestTopic,
		replyTopic:   replyTopic,
		error:        make(chan error, 1),
		stop:         make(chan struct{}),
		calls:        make(map[string]*pendingCall),
		callTimeout:  _defaultCallTimeout,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(c)
	}

	err := c.conn.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("kafka_rpc client - NewClient - c.conn.Connect: %w", err)
	}

	// Subscribe to reply topic
	c.conn.Client.AddConsumeTopics(c.replyTopic)

	go c.consumer()

	return c, nil
}

func (c *Client) publish(ctx context.Context, corrID, handler string, request interface{}) error {
	var (
		requestBody []byte
		err         error
	)

	if request != nil {
		requestBody, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}

	headers := []kgo.RecordHeader{
		{Key: "handler", Value: []byte(handler)},
		{Key: "correlation_id", Value: []byte(corrID)},
		{Key: "reply_topic", Value: []byte(c.replyTopic)},
	}

	record := &kgo.Record{
		Topic:   c.requestTopic,
		Key:     []byte(corrID),
		Value:   requestBody,
		Headers: headers,
	}

	results := c.conn.Client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("c.Client.ProduceSync: %w", err)
	}

	return nil
}

// RemoteCall performs a synchronous RPC call to a remote handler.
// It sends a request and waits for a response or timeout.
//
// Parameters:
//   - ctx: context for cancellation
//   - handler: the name of the remote handler to call
//   - request: the request payload (will be JSON marshaled)
//   - response: pointer to store the response (will be JSON unmarshaled)
//
// Returns an error if the call times out, the connection is closed,
// or the remote handler returns an error.
func (c *Client) RemoteCall(ctx context.Context, handler string, request, response interface{}) error {
	select {
	case <-c.stop:
		return ErrConnectionClosed
	default:
	}

	corrID := uuid.New().String()

	err := c.publish(ctx, corrID, handler, request)
	if err != nil {
		return fmt.Errorf("kafka_rpc client - Client - RemoteCall - c.publish: %w", err)
	}

	call := &pendingCall{done: make(chan struct{})}

	c.addCall(corrID, call)
	defer c.deleteCall(corrID)

	timeoutCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()

	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return kafka.ErrTimeout
		}
		return timeoutCtx.Err()
	case <-call.done:
	}

	if call.err != nil {
		return call.err
	}

	if call.status == kafka.Success {
		err = json.Unmarshal(call.body, response)
		if err != nil {
			return fmt.Errorf("kafka_rpc client - Client - RemoteCall - json.Unmarshal: %w", err)
		}
		return nil
	}

	if call.status == kafka.ErrBadHandler.Error() {
		return kafka.ErrBadHandler
	}

	if call.status == kafka.ErrInternalServer.Error() {
		return kafka.ErrInternalServer
	}

	return nil
}

func (c *Client) consumer() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}

		fetches := c.conn.Client.PollFetches(c.conn.Context())
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				select {
				case c.error <- err.Err:
				default:
				}
			}
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			c.handleResponse(record)
		})
	}
}

func (c *Client) handleResponse(record *kgo.Record) {
	var corrID string
	for _, header := range record.Headers {
		if header.Key == "correlation_id" {
			corrID = string(header.Value)
			break
		}
	}

	if corrID == "" {
		return
	}

	c.rw.RLock()
	call, ok := c.calls[corrID]
	c.rw.RUnlock()

	if !ok {
		return
	}

	// Extract status from headers
	status := kafka.Success
	for _, header := range record.Headers {
		if header.Key == "status" {
			status = string(header.Value)
			break
		}
	}

	call.status = status
	call.body = record.Value
	close(call.done)
}

func (c *Client) addCall(corrID string, call *pendingCall) {
	c.rw.Lock()
	c.calls[corrID] = call
	c.rw.Unlock()
}

func (c *Client) deleteCall(corrID string) {
	c.rw.Lock()
	delete(c.calls, corrID)
	c.rw.Unlock()
}

// Notify returns a channel that receives connection errors.
// The channel is closed when a fatal error occurs that requires recreating the client.
func (c *Client) Notify() <-chan error {
	return c.error
}

// Shutdown gracefully closes the Kafka client connection.
// It stops consuming messages and closes the underlying connection.
// Returns an error if the connection close fails.
func (c *Client) Shutdown() error {
	select {
	case <-c.error:
		return nil
	default:
	}

	close(c.stop)
	c.conn.Close()

	return nil
}
