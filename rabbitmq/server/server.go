// Package server provides a RabbitMQ RPC server implementation for handling remote procedure calls.
package server

import (
	"fmt"
	"time"

	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rdashevsky/go-pkgs/logger"
	rmqrpc "github.com/rdashevsky/go-pkgs/rabbitmq"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// CallHandler is a function that processes an incoming RPC request.
// It receives the AMQP delivery containing the request and returns a response and/or error.
// The response will be JSON marshaled before sending back to the client.
type CallHandler func(*amqp.Delivery) (interface{}, error)

// Server represents a RabbitMQ RPC server that handles incoming requests.
// It manages the connection, routes requests to appropriate handlers,
// and sends responses back to clients.
type Server struct {
	conn   *rmqrpc.Connection
	error  chan error
	stop   chan struct{}
	router map[string]CallHandler

	timeout time.Duration

	logger logger.LoggerI
}

// New creates a new RabbitMQ RPC server with the specified configuration.
// The server establishes a connection immediately but does not start consuming until Start is called.
//
// Parameters:
//   - url: RabbitMQ connection URL (e.g., "amqp://guest:guest@localhost:5672/")
//   - serverExchange: exchange name where requests will be received
//   - router: map of handler names to handler functions
//   - l: logger interface for error logging
//   - opts: optional configuration functions (Timeout, ConnWaitTime, ConnAttempts)
//
// Returns an error if the connection cannot be established.
func New(url, serverExchange string, router map[string]CallHandler, l logger.LoggerI, opts ...Option) (*Server, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	s := &Server{
		conn:    rmqrpc.New(serverExchange, cfg),
		error:   make(chan error),
		stop:    make(chan struct{}),
		router:  router,
		timeout: _defaultTimeout,
		logger:  l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	err := s.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc server - NewServer - s.conn.AttemptConnect: %w", err)
	}

	return s, nil
}

// Start begins consuming messages from the configured exchange.
// The server processes incoming requests in a separate goroutine.
// Use Notify() to receive server lifecycle errors.
func (s *Server) Start() {
	go s.consumer()
}

func (s *Server) consumer() {
	for {
		select {
		case <-s.stop:
			return
		case d, opened := <-s.conn.Delivery:
			if !opened {
				s.reconnect()

				return
			}

			_ = d.Ack(false) //nolint:errcheck // don't need this

			s.serveCall(&d)
		}
	}
}

func (s *Server) serveCall(d *amqp.Delivery) {
	callHandler, ok := s.router[d.Type]
	if !ok {
		s.publish(d, nil, rmqrpc.ErrBadHandler.Error())

		return
	}

	response, err := callHandler(d)
	if err != nil {
		s.publish(d, nil, rmqrpc.ErrInternalServer.Error())

		s.logger.Error(err, "rmq_rpc server - Server - serveCall - callHandler")

		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - serveCall - json.Marshal")
	}

	s.publish(d, body, rmqrpc.Success)
}

func (s *Server) publish(d *amqp.Delivery, body []byte, status string) {
	err := s.conn.Channel.Publish(d.ReplyTo, "", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Type:          status,
			Body:          body,
		})
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - publish - s.conn.Channel.Publish")
	}
}

func (s *Server) reconnect() {
	close(s.stop)

	err := s.conn.AttemptConnect()
	if err != nil {
		s.error <- err
		close(s.error)

		return
	}

	s.stop = make(chan struct{})

	go s.consumer()
}

// Notify returns a channel that receives server errors.
// The channel is closed when a fatal error occurs that requires recreating the server.
func (s *Server) Notify() <-chan error {
	return s.error
}

// Shutdown gracefully stops the RabbitMQ server.
// It stops consuming messages, waits for the configured timeout period,
// and then closes the underlying connection.
// Returns an error if the connection close fails.
func (s *Server) Shutdown() error {
	select {
	case <-s.error:
		return nil
	default:
	}

	close(s.stop)
	time.Sleep(s.timeout)

	err := s.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.Connection.Close: %w", err)
	}

	return nil
}
