// Package server provides a Kafka RPC server implementation for handling remote procedure calls.
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	kafka "github.com/rdashevsky/go-pkgs/kafka"
	"github.com/rdashevsky/go-pkgs/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

// CallHandler is a function that processes an incoming RPC request.
// It receives the Kafka record containing the request and returns a response and/or error.
// The response will be JSON marshaled before sending back to the client.
type CallHandler func(*kgo.Record) (interface{}, error)

// Server represents a Kafka RPC server that handles incoming requests.
// It manages the connection, routes requests to appropriate handlers,
// and sends responses back to clients.
type Server struct {
	conn         *kafka.Connection
	requestTopic string
	error        chan error
	stop         chan struct{}
	router       map[string]CallHandler

	logger logger.LoggerI
}

// New creates a new Kafka RPC server with the specified configuration.
// The server establishes a connection immediately but does not start consuming until Start is called.
//
// Parameters:
//   - cfg: Kafka connection configuration
//   - requestTopic: topic name where requests will be received
//   - router: map of handler names to handler functions
//   - l: logger interface for error logging
//   - opts: optional configuration functions
//
// Returns an error if the connection cannot be established.
func New(cfg kafka.Config, requestTopic string, router map[string]CallHandler, l logger.LoggerI, opts ...Option) (*Server, error) {
	// Ensure we have a consumer group for requests
	if cfg.GroupID == "" {
		return nil, fmt.Errorf("kafka_rpc server - NewServer - GroupID is required for server")
	}

	conn := kafka.NewConnection(cfg)

	s := &Server{
		conn:         conn,
		requestTopic: requestTopic,
		error:        make(chan error, 1),
		stop:         make(chan struct{}),
		router:       router,
		logger:       l,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(s)
	}

	err := s.conn.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("kafka_rpc server - NewServer - s.conn.Connect: %w", err)
	}

	// Subscribe to request topic
	s.conn.Client.AddConsumeTopics(s.requestTopic)

	return s, nil
}

// Start begins consuming messages from the configured topic.
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
		default:
		}

		fetches := s.conn.Client.PollFetches(s.conn.Context())
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, err := range errs {
				select {
				case s.error <- err.Err:
				default:
				}
			}
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			s.serveCall(record)
		})
	}
}

func (s *Server) serveCall(record *kgo.Record) {
	// Extract handler name and correlation ID from headers
	var handler, corrID, replyTopic string
	for _, header := range record.Headers {
		switch header.Key {
		case "handler":
			handler = string(header.Value)
		case "correlation_id":
			corrID = string(header.Value)
		case "reply_topic":
			replyTopic = string(header.Value)
		}
	}

	if handler == "" || corrID == "" || replyTopic == "" {
		s.logger.Error("kafka_rpc server - Server - serveCall - missing required headers",
			"handler", handler, "corrID", corrID, "replyTopic", replyTopic)
		return
	}

	callHandler, ok := s.router[handler]
	if !ok {
		s.publish(replyTopic, corrID, nil, kafka.ErrBadHandler.Error())
		return
	}

	response, err := callHandler(record)
	if err != nil {
		s.publish(replyTopic, corrID, nil, kafka.ErrInternalServer.Error())
		s.logger.Error(err, "kafka_rpc server - Server - serveCall - callHandler")
		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		s.logger.Error(err, "kafka_rpc server - Server - serveCall - json.Marshal")
		s.publish(replyTopic, corrID, nil, kafka.ErrInternalServer.Error())
		return
	}

	s.publish(replyTopic, corrID, body, kafka.Success)
}

func (s *Server) publish(replyTopic, corrID string, body []byte, status string) {
	headers := []kgo.RecordHeader{
		{Key: "correlation_id", Value: []byte(corrID)},
		{Key: "status", Value: []byte(status)},
	}

	record := &kgo.Record{
		Topic:   replyTopic,
		Key:     []byte(corrID),
		Value:   body,
		Headers: headers,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results := s.conn.Client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		s.logger.Error(err, "kafka_rpc server - Server - publish - s.conn.Client.ProduceSync")
	}
}

// Notify returns a channel that receives server errors.
// The channel is closed when a fatal error occurs that requires recreating the server.
func (s *Server) Notify() <-chan error {
	return s.error
}

// Shutdown gracefully stops the Kafka server.
// It stops consuming messages and closes the underlying connection.
// Returns an error if the connection close fails.
func (s *Server) Shutdown() error {
	select {
	case <-s.error:
		return nil
	default:
	}

	close(s.stop)
	s.conn.Close()

	return nil
}
