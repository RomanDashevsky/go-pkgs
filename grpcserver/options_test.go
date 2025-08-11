package grpcserver

import (
	"testing"
)

func TestPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "numeric port",
			port:     "8080",
			expected: ":8080",
		},
		{
			name:     "high port number",
			port:     "65535",
			expected: ":65535",
		},
		{
			name:     "low port number",
			port:     "80",
			expected: ":80",
		},
		{
			name:     "port with leading zeros",
			port:     "0080",
			expected: ":0080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				address: _defaultAddr,
			}

			option := Port(tt.port)
			option(server)

			if server.address != tt.expected {
				t.Errorf("Port(%q) set address to %q, expected %q",
					tt.port, server.address, tt.expected)
			}
		})
	}
}

func TestMultipleOptions(t *testing.T) {
	t.Run("last port wins", func(t *testing.T) {
		server := &Server{
			address: _defaultAddr,
		}

		// Apply multiple port options
		options := []Option{
			Port("8080"),
			Port("9090"),
			Port("3000"),
		}

		for _, opt := range options {
			opt(server)
		}

		// The last option should win
		if server.address != ":3000" {
			t.Errorf("expected address :3000, got %s", server.address)
		}
	})
}

func TestOptionFunction(t *testing.T) {
	t.Run("option is a function", func(t *testing.T) {
		opt := Port("8080")

		// Verify that Port returns a function
		if opt == nil {
			t.Fatal("Port should return a non-nil Option function")
		}

		// Test that the function can be called
		server := &Server{
			address: _defaultAddr,
		}
		opt(server)

		if server.address != ":8080" {
			t.Errorf("Option function did not set port correctly")
		}
	})
}
