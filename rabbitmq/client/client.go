// Package client provides a RabbitMQ RPC client implementation for making remote procedure calls.
package client

import (
	"errors"
	"fmt"
	"sync"
	"time"

	rmqrpc "github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ErrConnectionClosed is returned when attempting to make a remote call on a closed connection.
var ErrConnectionClosed = errors.New("rmq_rpc client - Client - RemoteCall - Connection closed")

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// Message represents a RabbitMQ message with all its properties.
// It can be used to construct messages for publishing or to inspect received messages.
type Message struct {
	Queue         string
	Priority      uint8
	ContentType   string
	Body          []byte
	ReplyTo       string
	CorrelationID string
}

type pendingCall struct {
	done   chan struct{}
	status string
	body   []byte
}

// Client represents a RabbitMQ RPC client for making remote procedure calls.
// It manages the connection, handles request-response correlation, and provides timeout support.
type Client struct {
	conn           *rmqrpc.Connection
	serverExchange string
	error          chan error
	stop           chan struct{}

	rw    sync.RWMutex
	calls map[string]*pendingCall

	timeout time.Duration
}

// New creates a new RabbitMQ RPC client with the specified configuration.
// The client establishes a connection immediately and starts consuming responses.
//
// Parameters:
//   - url: RabbitMQ connection URL (e.g., "amqp://guest:guest@localhost:5672/")
//   - serverExchange: exchange name where requests will be published
//   - clientExchange: exchange name where responses will be received
//   - opts: optional configuration functions (Timeout, ConnWaitTime, ConnAttempts)
//
// Returns an error if the connection cannot be established.
func New(url, serverExchange, clientExchange string, opts ...Option) (*Client, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	c := &Client{
		conn:           rmqrpc.New(clientExchange, cfg),
		serverExchange: serverExchange,
		error:          make(chan error),
		stop:           make(chan struct{}),
		calls:          make(map[string]*pendingCall),
		timeout:        _defaultTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(c)
	}

	err := c.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc client - NewClient - c.conn.AttemptConnect: %w", err)
	}

	go c.consumer()

	return c, nil
}

func (c *Client) publish(corrID, handler string, request interface{}) error {
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

	err = c.conn.Channel.Publish(c.serverExchange, "", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       c.conn.ConsumerExchange,
			Type:          handler,
			Body:          requestBody,
		})
	if err != nil {
		return fmt.Errorf("c.Channel.Publish: %w", err)
	}

	return nil
}

// RemoteCall performs a synchronous RPC call to a remote handler.
// It sends a request and waits for a response or timeout.
//
// Parameters:
//   - handler: the name of the remote handler to call
//   - request: the request payload (will be JSON marshaled)
//   - response: pointer to store the response (will be JSON unmarshaled)
//
// Returns an error if the call times out, the connection is closed,
// or the remote handler returns an error.
func (c *Client) RemoteCall(handler string, request, response interface{}) error { //nolint:cyclop // complex func
	select {
	case <-c.stop:
		time.Sleep(c.timeout)
		select {
		case <-c.stop:
			return ErrConnectionClosed
		default:
		}
	default:
	}

	corrID := uuid.New().String()

	err := c.publish(corrID, handler, request)
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - RemoteCall - c.publish: %w", err)
	}

	call := &pendingCall{done: make(chan struct{})}

	c.addCall(corrID, call)
	defer c.deleteCall(corrID)

	select {
	case <-time.After(c.timeout):
		return rmqrpc.ErrTimeout
	case <-call.done:
	}

	if call.status == rmqrpc.Success {
		err = json.Unmarshal(call.body, &response)
		if err != nil {
			return fmt.Errorf("rmq_rpc client - Client - RemoteCall - json.Unmarshal: %w", err)
		}

		return nil
	}

	if call.status == rmqrpc.ErrBadHandler.Error() {
		return rmqrpc.ErrBadHandler
	}

	if call.status == rmqrpc.ErrInternalServer.Error() {
		return rmqrpc.ErrInternalServer
	}

	return nil
}

func (c *Client) consumer() {
	for {
		select {
		case <-c.stop:
			return
		case d, opened := <-c.conn.Delivery:
			if !opened {
				c.reconnect()

				return
			}

			_ = d.Ack(false) //nolint:errcheck // don't need this

			c.getCall(&d)
		}
	}
}

func (c *Client) reconnect() {
	close(c.stop)

	err := c.conn.AttemptConnect()
	if err != nil {
		c.error <- err
		close(c.error)

		return
	}

	c.stop = make(chan struct{})

	go c.consumer()
}

func (c *Client) getCall(d *amqp.Delivery) {
	c.rw.RLock()
	call, ok := c.calls[d.CorrelationId]
	c.rw.RUnlock()

	if !ok {
		return
	}

	call.status = d.Type
	call.body = d.Body
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

// Shutdown gracefully closes the RabbitMQ client connection.
// It waits for the configured timeout period before closing the underlying connection.
// Returns an error if the connection close fails.
func (c *Client) Shutdown() error {
	select {
	case <-c.error:
		return nil
	default:
	}

	close(c.stop)
	time.Sleep(c.timeout)

	err := c.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - Shutdown - c.Connection.Close: %w", err)
	}

	return nil
}
