// Package httpserver provides HTTP server utilities based on Fiber framework.
// It offers configurable timeouts, prefork mode, and graceful shutdown capabilities.
package httpserver

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

const (
	_defaultAddr            = ":80"
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultShutdownTimeout = 3 * time.Second
)

// Server represents an HTTP server with configurable options.
// It wraps Fiber application with additional features like graceful shutdown.
type Server struct {
	// App is the underlying Fiber application instance.
	App    *fiber.App
	notify chan error

	address         string
	prefork         bool
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration
}

// New creates a new HTTP server with the given options.
// Default configuration: port :80, read/write timeout 5s, shutdown timeout 3s.
//
// Example:
//
//	server := httpserver.New(
//	    httpserver.Port(":8080"),
//	    httpserver.ReadTimeout(10 * time.Second),
//	)
func New(opts ...Option) *Server {
	s := &Server{
		App:             nil,
		notify:          make(chan error, 1),
		address:         _defaultAddr,
		readTimeout:     _defaultReadTimeout,
		writeTimeout:    _defaultWriteTimeout,
		shutdownTimeout: _defaultShutdownTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	app := fiber.New(fiber.Config{
		Prefork:      s.prefork,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		JSONDecoder:  json.Unmarshal,
		JSONEncoder:  json.Marshal,
	})

	s.App = app

	return s
}

// Start begins listening for HTTP requests in a separate goroutine.
// Use Notify() to wait for startup errors or shutdown completion.
func (s *Server) Start() {
	go func() {
		s.notify <- s.App.Listen(s.address)
		close(s.notify)
	}()
}

// Notify returns a channel that will receive an error if the server
// fails to start or when the server shuts down.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully shuts down the server within the configured timeout.
func (s *Server) Shutdown() error {
	return s.App.ShutdownWithTimeout(s.shutdownTimeout)
}
