package client

import "time"

// Option is a function that configures a Client.
type Option func(*Client)

// CallTimeout sets the timeout for individual RPC calls.
func CallTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.callTimeout = timeout
	}
}
