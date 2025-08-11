package redis_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rdashevsky/go-pkgs/redis"
)

// Example demonstrates creating and using a Redis client
func Example() {
	// Create client with default TTL
	client, err := redis.New("localhost:6379", "", "", redis.TTL(5*time.Minute))
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Set a value with default TTL
	err = client.Set(ctx, "user:123", "john_doe")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Set a value with custom TTL
	err = client.SetWithTTL(ctx, "session:abc", "active", 1*time.Hour)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Get a value
	value, err := client.Get(ctx, "user:123")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	_ = value // Use the retrieved value
	fmt.Println("Redis client operations completed successfully")
}

// ExampleNew demonstrates different ways to create a Redis client
func ExampleNew() {
	// Basic client without authentication
	client1, err := redis.New("localhost:6379", "", "")
	if err != nil {
		fmt.Printf("Failed to create basic client: %v\n", err)
		return
	}
	defer client1.Close()

	// Client with authentication
	client2, err := redis.New("localhost:6379", "username", "password")
	if err != nil {
		fmt.Printf("Failed to create authenticated client: %v\n", err)
		return
	}
	defer client2.Close()

	// Client with custom TTL
	client3, err := redis.New("localhost:6379", "", "", redis.TTL(10*time.Minute))
	if err != nil {
		fmt.Printf("Failed to create client with TTL: %v\n", err)
		return
	}
	defer client3.Close()

	fmt.Println("Redis clients created successfully")
}

// ExampleRedis_Set demonstrates basic key-value storage
func ExampleRedis_Set() {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Set a simple key-value pair
	err = client.Set(ctx, "greeting", "Hello, World!")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Println("Value stored successfully")
}

// ExampleRedis_SetWithTTL demonstrates storing data with custom expiration
func ExampleRedis_SetWithTTL() {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Store a session token that expires in 30 minutes
	err = client.SetWithTTL(ctx, "session:user123", "abc123token", 30*time.Minute)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Store a cache entry that expires in 5 minutes
	err = client.SetWithTTL(ctx, "cache:expensive_computation", "result_data", 5*time.Minute)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Println("Values stored with TTL successfully")
}

// ExampleRedis_Get demonstrates retrieving stored values
func ExampleRedis_Get() {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// First, store a value
	err = client.Set(ctx, "user:name", "Alice")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Retrieve the value
	name, err := client.Get(ctx, "user:name")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Printf("Retrieved name: %s\n", name)

	// Attempt to get a non-existent key
	missing, err := client.Get(ctx, "non:existent")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Printf("Non-existent key returns: '%s' (empty string)\n", missing)
}

// ExampleTTL demonstrates using the TTL option
func ExampleTTL() {
	// Create client with 1-hour default TTL
	client, err := redis.New(
		"localhost:6379",
		"",
		"",
		redis.TTL(1*time.Hour),
	)
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// This Set operation will use the 1-hour TTL by default
	err = client.Set(ctx, "long_lived_data", "important_info")
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// This SetWithTTL operation will use a custom TTL, overriding the default
	err = client.SetWithTTL(ctx, "short_lived_data", "temp_info", 30*time.Second)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Println("Data stored with different TTLs")
}

// Example_userSession demonstrates a realistic user session management scenario
func Example_userSession() {
	// Create client for session management with 30-minute default TTL
	sessionClient, err := redis.New(
		"localhost:6379",
		"",
		"",
		redis.TTL(30*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer sessionClient.Close()

	ctx := context.Background()
	userID := "user123"
	sessionToken := "session_abc123"

	// Store user session
	sessionKey := fmt.Sprintf("session:%s", userID)
	err = sessionClient.Set(ctx, sessionKey, sessionToken)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Store user preferences with longer TTL
	prefsKey := fmt.Sprintf("prefs:%s", userID)
	err = sessionClient.SetWithTTL(ctx, prefsKey, "theme=dark,lang=en", 7*24*time.Hour) // 1 week
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	// Retrieve session to validate user
	retrievedSession, err := sessionClient.Get(ctx, sessionKey)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	if retrievedSession == sessionToken {
		fmt.Println("User session valid")
	}

	// Retrieve user preferences
	prefs, err := sessionClient.Get(ctx, prefsKey)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Printf("User preferences: %s\n", prefs)
}

// Example_caching demonstrates using Redis for application caching
func Example_caching() {
	// Create client for caching with 10-minute default TTL
	cacheClient, err := redis.New(
		"localhost:6379",
		"",
		"",
		redis.TTL(10*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer cacheClient.Close()

	ctx := context.Background()

	// Simulate caching expensive computation results
	cacheKey := "computation:result:12345"

	// Check if result is cached
	cachedResult, err := cacheClient.Get(ctx, cacheKey)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	if cachedResult != "" {
		fmt.Printf("Using cached result: %s\n", cachedResult)
	} else {
		// Simulate expensive computation
		computedResult := "expensive_computation_result"

		// Cache the result for future use
		err = cacheClient.Set(ctx, cacheKey, computedResult)
		if err != nil {
			fmt.Printf("Redis not available for example: %v\n", err)
			return
		}

		fmt.Printf("Computed and cached result: %s\n", computedResult)
	}

	// Cache frequently accessed data with longer TTL
	frequentDataKey := "frequent:data:config"
	err = cacheClient.SetWithTTL(ctx, frequentDataKey, "config_data_here", 1*time.Hour)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	fmt.Println("Caching operations completed")
}

// Example_rateLimiting demonstrates using Redis for rate limiting
func Example_rateLimiting() {
	client, err := redis.New("localhost:6379", "", "")
	if err != nil {
		fmt.Printf("Failed to create Redis client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	userID := "user456"

	// Create rate limiting key (e.g., 10 requests per minute)
	rateLimitKey := fmt.Sprintf("rate_limit:%s:%d", userID, time.Now().Unix()/60)

	// Check current request count
	currentCount, err := client.Get(ctx, rateLimitKey)
	if err != nil {
		fmt.Printf("Redis not available for example: %v\n", err)
		return
	}

	if currentCount == "" {
		// First request in this time window
		err = client.SetWithTTL(ctx, rateLimitKey, "1", 1*time.Minute)
		if err != nil {
			fmt.Printf("Redis not available for example: %v\n", err)
			return
		}
		fmt.Println("Request allowed (first in window)")
	} else {
		// Subsequent requests would increment the counter
		// This is a simplified example - real implementation would use INCR
		fmt.Printf("Current request count: %s\n", currentCount)
	}

	fmt.Println("Rate limiting check completed")
}

// Example_multipleClients demonstrates managing multiple Redis clients for different purposes
func Example_multipleClients() {
	// Client for session management
	sessionClient, err := redis.New(
		"localhost:6379",
		"",
		"",
		redis.TTL(30*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create session Redis client: %v\n", err)
		return
	}
	defer sessionClient.Close()

	// Client for caching with shorter TTL
	cacheClient, err := redis.New(
		"localhost:6379",
		"",
		"",
		redis.TTL(5*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create cache Redis client: %v\n", err)
		return
	}
	defer cacheClient.Close()

	ctx := context.Background()

	// Use session client for user sessions
	err = sessionClient.Set(ctx, "session:user789", "session_token_789")
	if err != nil {
		fmt.Printf("Redis not available for session example: %v\n", err)
		return
	}

	// Use cache client for temporary data
	err = cacheClient.Set(ctx, "cache:api_response", "cached_api_data")
	if err != nil {
		fmt.Printf("Redis not available for cache example: %v\n", err)
		return
	}

	fmt.Println("Multiple clients configured and used successfully")
}
