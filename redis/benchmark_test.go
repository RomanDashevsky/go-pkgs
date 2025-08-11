package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/redis"
)

// BenchmarkRedis_New benchmarks client creation
func BenchmarkRedis_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client, _ := redis.New("localhost:6379", "", "")
		if client != nil {
			client.Close()
		}
	}
}

// BenchmarkRedis_NewWithOptions benchmarks client creation with options
func BenchmarkRedis_NewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client, _ := redis.New(
			"localhost:6379",
			"user",
			"password",
			redis.TTL(10*time.Minute),
		)
		if client != nil {
			client.Close()
		}
	}
}

// BenchmarkRedis_NewWithCredentials benchmarks client creation with authentication
func BenchmarkRedis_NewWithCredentials(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client, _ := redis.New("localhost:6379", "testuser", "testpass")
		if client != nil {
			client.Close()
		}
	}
}

// BenchmarkRedis_NewWithMultipleTTLOptions benchmarks client creation with various TTL values
func BenchmarkRedis_NewWithMultipleTTLOptions(b *testing.B) {
	ttlValues := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ttl := ttlValues[i%len(ttlValues)]
		client, _ := redis.New("localhost:6379", "", "", redis.TTL(ttl))
		if client != nil {
			client.Close()
		}
	}
}

// BenchmarkRedis_Close benchmarks client cleanup
func BenchmarkRedis_Close(b *testing.B) {
	clients := make([]*redis.Redis, b.N)
	for i := 0; i < b.N; i++ {
		client, _ := redis.New("localhost:6379", "", "")
		clients[i] = client
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if clients[i] != nil {
			clients[i].Close()
		}
	}
}

// BenchmarkRedis_Set benchmarks Set operations (requires running Redis instance)
func BenchmarkRedis_Set(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark-key-" + string(rune(i))
		value := "benchmark-value-" + string(rune(i))
		err := client.Set(ctx, key, value)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_SetWithTTL benchmarks SetWithTTL operations (requires running Redis instance)
func BenchmarkRedis_SetWithTTL(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()
	ttl := 5 * time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark-ttl-key-" + string(rune(i))
		value := "benchmark-ttl-value-" + string(rune(i))
		err := client.SetWithTTL(ctx, key, value, ttl)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_Get benchmarks Get operations (requires running Redis instance)
func BenchmarkRedis_Get(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	// Pre-populate some keys
	for i := 0; i < 100; i++ {
		key := "benchmark-get-key-" + string(rune(i))
		value := "benchmark-get-value-" + string(rune(i))
		err := client.Set(ctx, key, value)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark-get-key-" + string(rune(i%100))
		_, err := client.Get(ctx, key)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_GetNonExistent benchmarks Get operations for non-existent keys
func BenchmarkRedis_GetNonExistent(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "non-existent-key-" + string(rune(i))
		_, err := client.Get(ctx, key)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_SetGetCycle benchmarks a complete set-then-get cycle
func BenchmarkRedis_SetGetCycle(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "cycle-key-" + string(rune(i))
		value := "cycle-value-" + string(rune(i))

		// Set
		err := client.Set(ctx, key, value)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}

		// Get
		_, err = client.Get(ctx, key)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_TTLVariations benchmarks SetWithTTL with different TTL durations
func BenchmarkRedis_TTLVariations(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()
	ttlValues := []time.Duration{
		1 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		1 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "ttl-variation-key-" + string(rune(i))
		value := "ttl-variation-value-" + string(rune(i))
		ttl := ttlValues[i%len(ttlValues)]

		err := client.SetWithTTL(ctx, key, value, ttl)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}
}

// BenchmarkRedis_ConcurrentSet benchmarks concurrent Set operations
func BenchmarkRedis_ConcurrentSet(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "concurrent-set-key-" + string(rune(i))
			value := "concurrent-set-value-" + string(rune(i))
			err := client.Set(ctx, key, value)
			if err != nil {
				b.Skip("Redis server not available for benchmark")
			}
			i++
		}
	})
}

// BenchmarkRedis_ConcurrentGet benchmarks concurrent Get operations
func BenchmarkRedis_ConcurrentGet(b *testing.B) {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		b.Skip("Redis server not available for benchmark")
	}
	defer client.Close()

	ctx := context.Background()

	// Pre-populate some keys
	for i := 0; i < 1000; i++ {
		key := "concurrent-get-key-" + string(rune(i))
		value := "concurrent-get-value-" + string(rune(i))
		err := client.Set(ctx, key, value)
		if err != nil {
			b.Skip("Redis server not available for benchmark")
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "concurrent-get-key-" + string(rune(i%1000))
			_, err := client.Get(ctx, key)
			if err != nil {
				b.Skip("Redis server not available for benchmark")
			}
			i++
		}
	})
}
