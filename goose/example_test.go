package goose_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rdashevsky/go-pkgs/goose"
	"github.com/rdashevsky/go-pkgs/logger"
)

// Example demonstrates basic usage of CheckMigrationStatus function
func Example() {
	// Create a PostgreSQL connection pool
	config, err := pgxpool.ParseConfig("postgres://user:password@localhost:5432/database?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Create a logger
	logger := logger.New("info")

	// Check if database is at expected migration version
	expectedVersion := int64(10)
	currentVersion, err := goose.CheckMigrationStatus(pool, expectedVersion, logger)
	if err != nil {
		fmt.Printf("Migration check failed: %v\n", err)
		fmt.Printf("Current version: %d, Expected: %d\n", currentVersion, expectedVersion)
		return
	}

	fmt.Printf("Database is up to date at version %d\n", currentVersion)
}

// ExampleCheckMigrationStatus_basicUsage demonstrates the most common use case
func ExampleCheckMigrationStatus_basicUsage() {
	// Note: This example will not run successfully without a real database connection
	config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/myapp")
	if err != nil {
		log.Printf("Config error: %v", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("Pool creation error: %v", err)
		return
	}
	defer pool.Close()

	logger := logger.New("info")
	expectedVersion := int64(5)

	version, err := goose.CheckMigrationStatus(pool, expectedVersion, logger)
	if err != nil {
		log.Printf("Migration status check failed: %v", err)
		log.Printf("Current database version: %d, expected: %d", version, expectedVersion)
		return
	}

	fmt.Printf("Migration check passed - database at version %d", version)
	// Output when successful:
	// Migration check passed - database at version 5
}

// ExampleCheckMigrationStatus_withTimeout demonstrates using context timeout
func ExampleCheckMigrationStatus_withTimeout() {
	// Create connection with timeout
	config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/myapp?connect_timeout=5")
	if err != nil {
		log.Printf("Config error: %v", err)
		return
	}

	// Set additional pool configuration
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Pool creation error: %v", err)
		return
	}
	defer pool.Close()

	logger := logger.New("debug")
	expectedVersion := int64(15)

	version, err := goose.CheckMigrationStatus(pool, expectedVersion, logger)
	if err != nil {
		fmt.Printf("Error: %v (current version: %d)", err, version)
		return
	}

	fmt.Printf("Success: database at version %d", version)
}

// ExampleCheckMigrationStatus_errorHandling demonstrates comprehensive error handling
func ExampleCheckMigrationStatus_errorHandling() {
	config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/myapp")
	if err != nil {
		fmt.Printf("Configuration error: %v", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Printf("Connection error: %v", err)
		return
	}
	defer pool.Close()

	logger := logger.New("warn")
	expectedVersion := int64(20)

	version, err := goose.CheckMigrationStatus(pool, expectedVersion, logger)

	switch {
	case err == nil:
		fmt.Printf("✓ Database schema is up to date at version %d", version)
	case version == 0:
		fmt.Printf("✗ Failed to connect to database or retrieve version: %v", err)
	case version != expectedVersion:
		fmt.Printf("✗ Schema version mismatch: found %d, expected %d", version, expectedVersion)
		fmt.Printf("  Please run database migrations to update to version %d", expectedVersion)
	default:
		fmt.Printf("✗ Unexpected error: %v", err)
	}
}

// ExampleCheckMigrationStatus_developmentSetup shows a typical development setup
func ExampleCheckMigrationStatus_developmentSetup() {
	// Development database configuration
	dsn := "postgres://dev:devpass@localhost:5432/myapp_dev?sslmode=disable"

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Sprintf("Invalid database config: %v", err))
	}

	// Development-friendly pool settings
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(fmt.Sprintf("Cannot connect to development database: %v", err))
	}
	defer pool.Close()

	// Use debug logging in development
	logger := logger.New("debug")

	// Check for latest migration (example: version 12)
	latestVersion := int64(12)

	currentVersion, err := goose.CheckMigrationStatus(pool, latestVersion, logger)
	if err != nil {
		fmt.Printf("⚠️  Migration check failed!")
		fmt.Printf("   Current version: %d", currentVersion)
		fmt.Printf("   Expected version: %d", latestVersion)
		fmt.Printf("   Error: %v", err)
		fmt.Printf("   Run: goose up")
		return
	}

	fmt.Printf("✅ Development database ready (schema version %d)", currentVersion)
}

// ExampleCheckMigrationStatus_productionSetup shows a production-ready setup
func ExampleCheckMigrationStatus_productionSetup() {
	// Production database with SSL and connection limits
	dsn := "postgres://produser:securepass@prod-db:5432/myapp_prod?sslmode=require&pool_max_conns=20"

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Production database config error: %v", err)
	}

	// Production pool settings
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Shorter timeout for production health checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		cancel()
		log.Fatalf("Cannot connect to production database: %v", err)
	}
	defer func() {
		pool.Close()
		cancel()
	}()

	// Use error-level logging in production to reduce noise
	logger := logger.New("error")

	// Production migration version
	requiredVersion := int64(25)

	currentVersion, err := goose.CheckMigrationStatus(pool, requiredVersion, logger)
	if err != nil {
		log.Printf("CRITICAL: Production database schema mismatch!")
		log.Printf("Current: %d, Required: %d, Error: %v", currentVersion, requiredVersion, err)
		// In production, you might want to fail fast or alert monitoring systems
		return
	}

	log.Printf("Production database ready (schema v%d)", currentVersion)
}

// ExampleCheckMigrationStatus_multipleEnvironments shows environment-specific handling
func ExampleCheckMigrationStatus_multipleEnvironments() {
	environments := map[string]struct {
		dsn             string
		expectedVersion int64
		logLevel        string
	}{
		"development": {
			dsn:             "postgres://dev:dev@localhost:5432/myapp_dev?sslmode=disable",
			expectedVersion: 10,
			logLevel:        "debug",
		},
		"staging": {
			dsn:             "postgres://stage:stage@staging-db:5432/myapp_stage?sslmode=require",
			expectedVersion: 15,
			logLevel:        "info",
		},
		"production": {
			dsn:             "postgres://prod:securepass@prod-db:5432/myapp_prod?sslmode=require",
			expectedVersion: 20,
			logLevel:        "error",
		},
	}

	env := "development" // This would typically come from environment variable
	config := environments[env]

	poolConfig, err := pgxpool.ParseConfig(config.dsn)
	if err != nil {
		fmt.Printf("Environment %s: config error: %v", env, err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		fmt.Printf("Environment %s: connection error: %v", env, err)
		return
	}
	defer pool.Close()

	logger := logger.New(config.logLevel)

	version, err := goose.CheckMigrationStatus(pool, config.expectedVersion, logger)
	if err != nil {
		fmt.Printf("Environment %s: migration check failed (current: %d, expected: %d): %v",
			env, version, config.expectedVersion, err)
		return
	}

	fmt.Printf("Environment %s: ready at schema version %d", env, version)
}

// ExampleCheckMigrationStatus_healthCheck shows using the function for health checks
func ExampleCheckMigrationStatus_healthCheck() {
	config, err := pgxpool.ParseConfig("postgres://health:check@localhost:5432/myapp")
	if err != nil {
		fmt.Printf("Health check config error: %v", err)
		return
	}

	// Quick timeout for health checks
	config.ConnConfig.ConnectTimeout = 2 * time.Second
	config.MaxConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Printf("Health check: database connection failed: %v", err)
		return
	}
	defer pool.Close()

	// Minimal logging for health checks
	logger := &benchmarkLogger{} // Use no-op logger

	expectedVersion := int64(18)
	version, err := goose.CheckMigrationStatus(pool, expectedVersion, logger)

	// Health check response
	if err != nil {
		fmt.Printf("UNHEALTHY: schema version %d (expected %d): %v", version, expectedVersion, err)
		return
	}

	fmt.Printf("HEALTHY: database ready (schema v%d)", version)
	// Output when healthy:
	// HEALTHY: database ready (schema v18)
}

// customLogger is a custom logger implementation that prefixes all messages
type customLogger struct {
	prefix string
}

// Implement logger.LoggerI interface
func (c *customLogger) Debug(message interface{}, args ...interface{}) {
	fmt.Printf("[%s] DEBUG: %v", c.prefix, message)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

func (c *customLogger) Info(message string, args ...interface{}) {
	fmt.Printf("[%s] INFO: ", c.prefix)
	fmt.Printf(message, args...)
	fmt.Println()
}

func (c *customLogger) Warn(message string, args ...interface{}) {
	fmt.Printf("[%s] WARN: ", c.prefix)
	fmt.Printf(message, args...)
	fmt.Println()
}

func (c *customLogger) Error(message interface{}, args ...interface{}) {
	fmt.Printf("[%s] ERROR: %v", c.prefix, message)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

func (c *customLogger) Fatal(message interface{}, args ...interface{}) {
	fmt.Printf("[%s] FATAL: %v", c.prefix, message)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

// Example_customLogger demonstrates using a custom logger implementation
func Example_customLogger() {
	// Usage with custom logger
	config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/myapp")
	if err != nil {
		log.Printf("Config error: %v", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("Pool creation error: %v", err)
		return
	}
	defer pool.Close()

	customLog := &customLogger{prefix: "MIGRATION-CHECK"}
	expectedVersion := int64(8)

	version, err := goose.CheckMigrationStatus(pool, expectedVersion, customLog)
	if err != nil {
		fmt.Printf("Check failed: %v (version: %d)", err, version)
		return
	}

	fmt.Printf("Success with custom logger: version %d", version)
}
