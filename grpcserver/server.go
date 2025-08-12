// Package grpcserver provides a simple gRPC server implementation with graceful shutdown support.
package grpcserver

import (
	"fmt"
	"net"

	pbgrpc "google.golang.org/grpc"
)

const (
	_defaultAddr = ":80"
)

// Server represents a gRPC server with lifecycle management.
// It wraps google.golang.org/grpc.Server with additional functionality
// for monitoring server state and graceful shutdown.
type Server struct {
	App     *pbgrpc.Server
	notify  chan error
	address string
}

// New creates a new gRPC server instance with the specified options.
// By default, the server listens on port 80. Use Port option to customize.
//
// Example:
//
//	server := grpcserver.New(grpcserver.Port("8080"))
//	grpc_health_v1.RegisterHealthServer(server.App, health.NewServer())
//	server.Start()
func New(opts ...Option) *Server {
	s := &Server{
		App:     pbgrpc.NewServer(),
		notify:  make(chan error, 1),
		address: _defaultAddr,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start begins listening for gRPC connections on the configured address.
// The server runs in a separate goroutine and errors are sent to the notify channel.
// Use Notify() to receive server lifecycle events.
func (s *Server) Start() {
	go func() {
		ln, err := net.Listen("tcp", s.address)
		if err != nil {
			s.notify <- fmt.Errorf("failed to listen: %w", err)
			close(s.notify)

			return
		}

		s.notify <- s.App.Serve(ln)
		close(s.notify)
	}()
}

// Notify returns a channel that receives server lifecycle errors.
// The channel is closed when the server stops.
// This channel will receive errors from server startup (e.g., port already in use)
// or from unexpected server termination.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully stops the gRPC server.
// It waits for all active connections to close before returning.
// Always returns nil as GracefulStop does not return errors.
func (s *Server) Shutdown() error {
	s.App.GracefulStop()

	return nil
}
