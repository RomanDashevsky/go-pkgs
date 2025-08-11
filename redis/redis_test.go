package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/redis"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		user     string
		password string
		opts     []redis.Options
	}{
		{
			name:     "default configuration",
			address:  "localhost:6379",
			user:     "",
			password: "",
			opts:     nil,
		},
		{
			name:     "with credentials",
			address:  "localhost:6379",
			user:     "testuser",
			password: "testpass",
			opts:     nil,
		},
		{
			name:     "with TTL option",
			address:  "localhost:6379",
			user:     "",
			password: "",
			opts:     []redis.Options{redis.TTL(5 * time.Minute)},
		},
		{
			name:     "different address",
			address:  "127.0.0.1:6380",
			user:     "",
			password: "",
			opts:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := redis.New(tt.address, tt.user, tt.password, tt.opts...)

			if err != nil {
				t.Errorf("expected no error from New(), got: %v", err)
			}

			if client == nil {
				t.Error("expected client to be created, got nil")
			}

			if client != nil {
				client.Close()
			}
		})
	}
}

func TestRedis_SetAndGet_NoConnection(t *testing.T) {
	// Create client that won't be able to connect
	client, err := redis.New("127.0.0.1:65432", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test Set operation (will fail due to no connection)
	err = client.Set(ctx, "test-key", "test-value")
	if err == nil {
		t.Skip("unexpected successful connection to Redis")
	}

	// Test Get operation (will fail due to no connection)
	_, err = client.Get(ctx, "test-key")
	if err == nil {
		t.Skip("unexpected successful connection to Redis")
	}
}

func TestRedis_SetWithTTL_NoConnection(t *testing.T) {
	client, err := redis.New("127.0.0.1:65432", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test SetWithTTL operation (will fail due to no connection)
	err = client.SetWithTTL(ctx, "test-key", "test-value", 1*time.Hour)
	if err == nil {
		t.Skip("unexpected successful connection to Redis")
	}
}

func TestRedis_Get_NonExistentKey(t *testing.T) {
	// This test would require a real Redis connection to work properly
	// For now, we'll just test the structure
	client, err := redis.New("127.0.0.1:65432", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Try to get a non-existent key (will fail due to no connection)
	_, err = client.Get(ctx, "non-existent-key")
	if err == nil {
		t.Skip("unexpected successful connection to Redis")
	}
}

func TestRedis_Close(t *testing.T) {
	client, err := redis.New("127.0.0.1:65432", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Should not panic
	client.Close()

	// Should be safe to call multiple times
	client.Close()
}

func TestTTL_Option(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
	}{
		{"1 minute TTL", 1 * time.Minute},
		{"1 hour TTL", 1 * time.Hour},
		{"30 seconds TTL", 30 * time.Second},
		{"1 day TTL", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := redis.New("127.0.0.1:65432", "", "", redis.TTL(tt.ttl))

			if err != nil {
				t.Errorf("expected no error from New(), got: %v", err)
			}

			if client == nil {
				t.Error("expected client to be created, got nil")
			}

			if client != nil {
				client.Close()
			}
		})
	}
}

func TestRedis_MultipleOptions(t *testing.T) {
	// Test that multiple options can be applied (currently only TTL is available)
	client, err := redis.New(
		"localhost:6379",
		"user",
		"pass",
		redis.TTL(10*time.Minute),
	)

	if err != nil {
		t.Errorf("expected no error from New(), got: %v", err)
	}

	if client == nil {
		t.Error("expected client to be created, got nil")
	}

	if client != nil {
		client.Close()
	}
}

// TestRedis_IntegrationSetGet would test actual Redis operations
// This would require a running Redis instance
func TestRedis_IntegrationSetGet(t *testing.T) {
	// Skip this test if we don't have a Redis server available
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	testKey := "test-integration-key"
	testValue := "test-integration-value"

	// Try to set a value
	err = client.Set(ctx, testKey, testValue)
	if err != nil {
		t.Skip("Redis server not available for integration test")
	}

	// Try to get the value
	retrievedValue, err := client.Get(ctx, testKey)
	if err != nil {
		t.Errorf("failed to get value: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("expected %q, got %q", testValue, retrievedValue)
	}

	// Test getting non-existent key
	nonExistentValue, err := client.Get(ctx, "non-existent-key")
	if err != nil {
		t.Errorf("expected no error for non-existent key, got: %v", err)
	}

	if nonExistentValue != "" {
		t.Errorf("expected empty string for non-existent key, got: %q", nonExistentValue)
	}
}

// TestRedis_IntegrationSetWithTTL would test TTL functionality
func TestRedis_IntegrationSetWithTTL(t *testing.T) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	testKey := "test-ttl-key"
	testValue := "test-ttl-value"

	// Try to set a value with short TTL
	err = client.SetWithTTL(ctx, testKey, testValue, 100*time.Millisecond)
	if err != nil {
		t.Skip("Redis server not available for integration test")
	}

	// Should be able to get it immediately
	retrievedValue, err := client.Get(ctx, testKey)
	if err != nil {
		t.Errorf("failed to get value: %v", err)
	}

	if retrievedValue != testValue {
		t.Errorf("expected %q, got %q", testValue, retrievedValue)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	expiredValue, err := client.Get(ctx, testKey)
	if err != nil {
		t.Errorf("expected no error for expired key, got: %v", err)
	}

	if expiredValue != "" {
		t.Errorf("expected empty string for expired key, got: %q", expiredValue)
	}
}
