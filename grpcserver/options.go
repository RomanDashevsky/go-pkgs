package grpcserver

import (
	"net"
)

// Option is a function that configures a Server.
// Options are applied in the order they are passed to New.
type Option func(*Server)

// Port sets the port on which the gRPC server will listen.
// The port should be a string representation of a valid port number.
//
// Example:
//
//	server := grpcserver.New(grpcserver.Port("9090"))
func Port(port string) Option {
	return func(s *Server) {
		s.address = net.JoinHostPort("", port)
	}
}
