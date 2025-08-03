package postgres

import "time"

// Option defines a function type for configuring Postgres instances.
type Option func(*Postgres)

// MaxPoolSize sets the maximum number of connections in the pool.
func MaxPoolSize(size int) Option {
	return func(c *Postgres) {
		c.maxPoolSize = size
	}
}

// ConnAttempts sets the number of connection attempts before giving up.
func ConnAttempts(attempts int) Option {
	return func(c *Postgres) {
		c.connAttempts = attempts
	}
}

// ConnTimeout sets the timeout duration between connection attempts.
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connTimeout = timeout
	}
}
