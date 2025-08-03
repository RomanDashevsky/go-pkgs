package httpserver

import (
	"net"
	"time"
)

// Option defines a function type for configuring Server instances.
type Option func(*Server)

// Port sets the server listening port.
// The port should include the colon prefix, e.g., ":8080".
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}

// Prefork enables or disables prefork mode for better performance.
// When enabled, the server will spawn multiple child processes.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

// ReadTimeout sets the maximum duration for reading the entire request.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout sets the maximum duration before timing out writes of the response.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout sets the maximum duration to wait for graceful shutdown.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
