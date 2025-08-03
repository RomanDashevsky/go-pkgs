package redis

import "time"

// Options defines a function type for configuring Redis instances.
type Options func(*Redis)

// TTL sets the default TTL (time-to-live) for Set operations.
func TTL(ttl time.Duration) Options {
	return func(c *Redis) {
		c.ttl = ttl
	}
}
