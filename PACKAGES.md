# Package Documentation

This document provides detailed information about each package in the repository.

## Table of Contents

- [Logger](#logger)
- [HTTP Server](#http-server)
- [PostgreSQL](#postgresql)
- [Redis](#redis)
- [Goose](#goose)

## Logger

The logger package provides a structured logging interface based on zerolog with configurable levels and automatic caller information.

### Features

- Structured JSON logging
- Configurable log levels (debug, info, warn, error, fatal)
- Automatic caller information
- High performance with zerolog
- Interface-based design for easy testing

### API Reference

#### Types

```go
type LoggerI interface {
    Debug(message interface{}, args ...interface{})
    Info(message string, args ...interface{})
    Warn(message string, args ...interface{})
    Error(message interface{}, args ...interface{})
    Fatal(message interface{}, args ...interface{})
}

type Logger struct {
    // Internal implementation
}
```

#### Functions

```go
func New(level string) LoggerI
```
Creates a new logger with the specified level. Supported levels: "debug", "info", "warn", "error", "fatal".

### Example Usage

```go
package main

import (
    "github.com/rdashevsky/go-pkgs/logger"
)

func main() {
    // Create logger with info level
    l := logger.New("info")
    
    // Log messages
    l.Info("Application started")
    l.Info("Processing user %s", "john")
    l.Warn("High memory usage: %d%%", 85)
    l.Error("Failed to connect to database")
    
    // Debug messages won't be shown with info level
    l.Debug("Debug message")
}
```

---

## HTTP Server

The httpserver package provides an HTTP server based on Fiber with configurable options, middleware support, and graceful shutdown.

### Features

- Based on high-performance Fiber framework
- Configurable timeouts and connection pooling
- Graceful shutdown capabilities
- Built-in middleware (logging, recovery)
- Standardized error responses
- Options pattern for configuration

### API Reference

#### Types

```go
type Server struct {
    App *fiber.App
    // Internal fields
}

type Option func(*Server)
```

#### Functions

```go
func New(opts ...Option) *Server
```
Creates a new HTTP server with the given options.

#### Options

```go
func Port(port string) Option
func Prefork(prefork bool) Option
func ReadTimeout(timeout time.Duration) Option
func WriteTimeout(timeout time.Duration) Option
func ShutdownTimeout(timeout time.Duration) Option
```

#### Methods

```go
func (s *Server) Start()
func (s *Server) Shutdown() error
func (s *Server) Notify() <-chan error
```

### Example Usage

```go
package main

import (
    "time"
    "github.com/rdashevsky/go-pkgs/httpserver"
    "github.com/rdashevsky/go-pkgs/httpserver/middleware"
    "github.com/rdashevsky/go-pkgs/logger"
)

func main() {
    // Create logger
    l := logger.New("info")
    
    // Create server with options
    server := httpserver.New(
        httpserver.Port(":8080"),
        httpserver.ReadTimeout(10 * time.Second),
        httpserver.WriteTimeout(10 * time.Second),
        httpserver.ShutdownTimeout(5 * time.Second),
    )
    
    // Add middleware
    server.App.Use(middleware.Logger(l))
    server.App.Use(middleware.Recovery(l))
    
    // Add routes
    server.App.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(map[string]string{"status": "ok"})
    })
    
    // Start server
    server.Start()
    
    // Wait for shutdown signal
    <-server.Notify()
}
```

### Middleware

#### Logger Middleware

```go
import "github.com/rdashevsky/go-pkgs/httpserver/middleware"

server.App.Use(middleware.Logger(logger))
```

Logs all HTTP requests with method, URL, status code, and response time.

#### Recovery Middleware

```go
server.App.Use(middleware.Recovery(logger))
```

Recovers from panics and logs stack traces.

#### Error Response Utilities

```go
import "github.com/rdashevsky/go-pkgs/httpserver/response"

// Send error response
return response.Error(c, 404, 
    response.ErrorMessage(&message),
    response.ErrorTitle("Custom Not Found"),
)
```

---

## PostgreSQL

The postgres package provides PostgreSQL database connectivity with connection pooling and query building capabilities using pgx driver and Squirrel query builder.

### Features

- Connection pooling with pgx
- Integrated Squirrel query builder
- Configurable retry logic
- Connection attempt management
- Thread-safe operations

### API Reference

#### Types

```go
type Postgres struct {
    Builder squirrel.StatementBuilderType
    Pool    *pgxpool.Pool
    // Internal fields
}

type Option func(*Postgres)
```

#### Functions

```go
func New(url string, opts ...Option) (*Postgres, error)
```
Creates a new PostgreSQL connection with retry logic.

#### Options

```go
func MaxPoolSize(size int) Option
func ConnAttempts(attempts int) Option
func ConnTimeout(timeout time.Duration) Option
```

#### Methods

```go
func (p *Postgres) Close()
```

### Example Usage

```go
package main

import (
    "context"
    "time"
    "github.com/rdashevsky/go-pkgs/postgres"
    "github.com/Masterminds/squirrel"
)

func main() {
    // Create connection
    pg, err := postgres.New(
        "postgres://user:pass@localhost:5432/db?sslmode=disable",
        postgres.MaxPoolSize(10),
        postgres.ConnAttempts(3),
        postgres.ConnTimeout(5 * time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer pg.Close()
    
    // Use Squirrel query builder
    query, args, err := pg.Builder.
        Select("id", "name", "email").
        From("users").
        Where(squirrel.Eq{"active": true}).
        OrderBy("created_at DESC").
        Limit(10).
        ToSql()
    if err != nil {
        panic(err)
    }
    
    // Execute query
    rows, err := pg.Pool.Query(context.Background(), query, args...)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    
    // Process results
    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            panic(err)
        }
        // Process user data
    }
}
```

---

## Redis

The redis package provides Redis client functionality with configurable TTL and simplified key-value operations using go-redis/v9.

### Features

- Based on go-redis/v9
- Configurable default TTL
- Simple key-value operations
- Connection management
- Context-aware operations

### API Reference

#### Types

```go
type Redis struct {
    // Internal fields
}

type Options func(*Redis)
```

#### Functions

```go
func New(address string, user string, password string, opts ...Options) (*Redis, error)
```
Creates a new Redis client with the given credentials and options.

#### Options

```go
func TTL(ttl time.Duration) Options
```
Sets the default TTL for Set operations.

#### Methods

```go
func (r *Redis) Set(ctx context.Context, key string, value string) error
func (r *Redis) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error
func (r *Redis) Get(ctx context.Context, key string) (string, error)
func (r *Redis) Close()
```

### Example Usage

```go
package main

import (
    "context"
    "time"
    "github.com/rdashevsky/go-pkgs/redis"
)

func main() {
    // Create Redis client
    r, err := redis.New(
        "localhost:6379",
        "",           // username
        "",           // password
        redis.TTL(5 * time.Minute),
    )
    if err != nil {
        panic(err)
    }
    defer r.Close()
    
    ctx := context.Background()
    
    // Set value with default TTL
    err = r.Set(ctx, "user:123", "john_doe")
    if err != nil {
        panic(err)
    }
    
    // Set value with custom TTL
    err = r.SetWithTTL(ctx, "session:abc", "active", 1*time.Hour)
    if err != nil {
        panic(err)
    }
    
    // Get value
    value, err := r.Get(ctx, "user:123")
    if err != nil {
        panic(err)
    }
    
    if value == "" {
        // Key doesn't exist or expired
    } else {
        // Process value
    }
}
```

---

## Goose

The goose package provides database migration utilities built on top of pressly/goose with migration status checking and validation.

### Features

- Migration status checking
- Version validation
- Logger integration
- Built on pressly/goose
- Error handling with detailed messages

### API Reference

#### Functions

```go
func CheckMigrationStatus(pool *pgxpool.Pool, expectedVersion int64, l logger.LoggerI) (int64, error)
```
Checks if the current database migration version matches the expected version. Returns the current version and an error if versions don't match.

### Example Usage

```go
package main

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rdashevsky/go-pkgs/goose"
    "github.com/rdashevsky/go-pkgs/logger"
)

func main() {
    // Create logger
    l := logger.New("info")
    
    // Create database connection
    pool, err := pgxpool.New(context.Background(), 
        "postgres://user:pass@localhost:5432/db?sslmode=disable")
    if err != nil {
        panic(err)
    }
    defer pool.Close()
    
    // Check migration status
    expectedVersion := int64(10)
    currentVersion, err := goose.CheckMigrationStatus(pool, expectedVersion, l)
    if err != nil {
        l.Error("Migration check failed: %v", err)
        l.Error("Current version: %d, Expected: %d", currentVersion, expectedVersion)
        // Handle migration mismatch
        return
    }
    
    l.Info("Database migrations are up to date at version %d", currentVersion)
}
```

---

## Common Patterns

### Error Handling

All packages follow consistent error handling patterns:

```go
// Always check errors
result, err := someOperation()
if err != nil {
    logger.Error("Operation failed: %v", err)
    return err
}

// Use wrapped errors for context
if err := db.Connect(); err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}
```

### Context Usage

Packages that support context use it consistently:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

value, err := redisClient.Get(ctx, "key")
```

### Configuration

All packages use the options pattern for configuration:

```go
client := packagename.New(
    packagename.Option1(value1),
    packagename.Option2(value2),
)
```

### Resource Management

Always close resources properly:

```go
client := redis.New(...)
defer client.Close()

pool := postgres.New(...)
defer pool.Close()
```

### Testing

Use interfaces for easy testing:

```go
type Service struct {
    logger logger.LoggerI
}

// In tests, use mock logger
service := &Service{
    logger: &MockLogger{},
}
```