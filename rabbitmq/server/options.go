package server

import "time"

// Option is a function that configures a Server.
// Options are applied in the order they are passed to New.
type Option func(*Server)

// Timeout sets the duration to wait before closing the connection during shutdown.
// This allows pending operations to complete gracefully.
// Default is 2 seconds.
//
// Example:
//
//	server.New(url, exchange, router, logger, server.Timeout(5*time.Second))
func Timeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// ConnWaitTime sets the duration to wait between connection attempts.
// This is used when the server tries to establish the initial connection.
// Default is 5 seconds.
//
// Example:
//
//	server.New(url, exchange, router, logger, server.ConnWaitTime(2*time.Second))
func ConnWaitTime(timeout time.Duration) Option {
	return func(s *Server) {
		s.conn.WaitTime = timeout
	}
}

// ConnAttempts sets the maximum number of connection attempts.
// If all attempts fail, New returns an error.
// Default is 10 attempts.
//
// Example:
//
//	server.New(url, exchange, router, logger, server.ConnAttempts(3))
func ConnAttempts(attempts int) Option {
	return func(s *Server) {
		s.conn.Attempts = attempts
	}
}
