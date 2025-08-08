// Package redis provides Redis client functionality with configurable TTL
// and simplified key-value operations using go-redis/v9.
package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 2 * time.Minute

// Redis represents a Redis client with configurable default TTL.
type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

// New creates a new Redis client with the given connection parameters and options.
// Default TTL is set to 2 minutes for all Set operations.
//
// Example:
//
//	client, err := redis.New("localhost:6379", "", "",
//	    redis.TTL(5 * time.Minute),
//	)
func New(address string, user string, password string, opts ...Options) (*Redis, error) {
	r := &Redis{
		ttl: defaultTTL,
	}

	for _, opt := range opts {
		opt(r)
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:     address,
		Username: user,
		Password: password,
	})

	return r, nil
}

// Set stores a key-value pair with the default TTL.
func (r *Redis) Set(ctx context.Context, key string, value string) error {
	return r.SetWithTTL(ctx, key, value, r.ttl)
}

// SetWithTTL stores a key-value pair with a custom TTL.
func (r *Redis) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves the value for the given key.
// Returns empty string and nil error if key doesn't exist.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return val, nil
}

// Close gracefully closes the Redis client connection.
func (r *Redis) Close() {
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			log.Printf("Error closing redis client: %s", err)
		}
	}
}
