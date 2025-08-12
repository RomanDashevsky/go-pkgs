package client

import "time"

// Option is a function that configures a Client.
// Options are applied in the order they are passed to New.
type Option func(*Client)

// Timeout sets the maximum duration to wait for a response to a remote call.
// If a response is not received within this duration, RemoteCall returns ErrTimeout.
// Default is 2 seconds.
//
// Example:
//
//	client.New(url, serverEx, clientEx, client.Timeout(5*time.Second))
func Timeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// ConnWaitTime sets the duration to wait between connection attempts.
// This is used when the client tries to establish the initial connection.
// Default is 5 seconds.
//
// Example:
//
//	client.New(url, serverEx, clientEx, client.ConnWaitTime(2*time.Second))
func ConnWaitTime(timeout time.Duration) Option {
	return func(c *Client) {
		c.conn.WaitTime = timeout
	}
}

// ConnAttempts sets the maximum number of connection attempts.
// If all attempts fail, New returns an error.
// Default is 10 attempts.
//
// Example:
//
//	client.New(url, serverEx, clientEx, client.ConnAttempts(3))
func ConnAttempts(attempts int) Option {
	return func(c *Client) {
		c.conn.Attempts = attempts
	}
}
