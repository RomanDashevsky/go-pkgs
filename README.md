# Go Packages

A repository of shared Go packages for reuse across microservices.

## Packages

### Logger
A structured logger based on zerolog with configurable levels.
```go
import "github.com/rdashevsky/go-pkgs/logger"

l := logger.New("info")
l.Info("Application started")
```

### HTTP Server  
An HTTP server based on Fiber with configurable options.
```go
import "github.com/rdashevsky/go-pkgs/httpserver"

server := httpserver.New(
    httpserver.Port(":8080"),
    httpserver.ReadTimeout(10 * time.Second),
)
```

### PostgreSQL
A PostgreSQL client with connection pooling and Squirrel query builder.
```go
import "github.com/rdashevsky/go-pkgs/postgres"

pg, err := postgres.New(databaseURL,
    postgres.MaxPoolSize(10),
    postgres.ConnAttempts(5),
)
```

### Redis
A Redis client with configurable TTL.
```go
import "github.com/rdashevsky/go-pkgs/redis"

r, err := redis.New("localhost:6379", "", "",
    redis.TTL(5 * time.Minute),
)
```

### Goose
Utilities for database migration management.
```go
import (
    "github.com/rdashevsky/go-pkgs/goose"
    "github.com/rdashevsky/go-pkgs/logger"
)

l := logger.New("info")
currentVersion, err := goose.CheckMigrationStatus(pool, expectedVersion, l)
if err != nil {
    l.Error("Migration check failed: %v (current: %d)", err, currentVersion)
}
```

### gRPC Server
A gRPC server with graceful shutdown and configurable options.
```go
import "github.com/rdashevsky/go-pkgs/grpcserver"

server := grpcserver.New(
    grpcserver.Port(":50051"),
    grpcserver.ReadTimeout(30 * time.Second),
)
```

### RabbitMQ
RabbitMQ RPC client and server implementation with automatic reconnection.
```go
import "github.com/rdashevsky/go-pkgs/rabbitmq/client"

// RPC Client
client, err := client.New(
    "amqp://guest:guest@localhost:5672/",
    "server-exchange",
    "client-exchange",
)

var response MyResponse
err = client.RemoteCall("handler-name", request, &response)
```

```go
import "github.com/rdashevsky/go-pkgs/rabbitmq/server"

// RPC Server
router := map[string]server.CallHandler{
    "greet": func(d *amqp.Delivery) (interface{}, error) {
        return "Hello World", nil
    },
}

server, err := server.New(
    "amqp://guest:guest@localhost:5672/",
    "server-exchange",
    router,
    logger,
)
server.Start()
```

### Kafka
Kafka RPC client and server implementation using franz-go with producer/consumer support.
```go
import (
    "github.com/rdashevsky/go-pkgs/kafka"
    "github.com/rdashevsky/go-pkgs/kafka/client"
)

// Kafka RPC Client
cfg := kafka.Config{
    Brokers:  []string{"localhost:9092"},
    ClientID: "my-app",
    GroupID:  "my-group",
}

client, err := client.New(cfg, "requests", "replies")

var response MyResponse
err = client.RemoteCall(ctx, "handler-name", request, &response)
```

```go
import (
    "github.com/rdashevsky/go-pkgs/kafka/server"
    "github.com/twmb/franz-go/pkg/kgo"
)

// Kafka RPC Server
router := map[string]server.CallHandler{
    "greet": func(record *kgo.Record) (interface{}, error) {
        return "Hello World", nil
    },
}

server, err := server.New(cfg, "requests", router, logger)
server.Start()
```

## Usage

1. Add the module to your `go.mod`:
```bash
go get github.com/rdashevsky/go-pkgs
```

2. Import the required packages:
```go
import (
    "github.com/rdashevsky/go-pkgs/logger"
    "github.com/rdashevsky/go-pkgs/httpserver"
)
```

## Features

- 🚀 **High Performance**: Optimized for production use
- 🔧 **Configurable**: Flexible options for each package
- 🛡️ **Secure**: Security-first design with vulnerability scanning
- 📊 **Observable**: Built-in logging and monitoring support
- 🧪 **Well Tested**: Comprehensive test suite with >90% coverage
- 📚 **Documented**: Complete documentation with examples

## Development

### Prerequisites
- Go 1.24 or later
- Make
- Docker (for integration tests)

### Quick Start
```bash
# Clone the repository
git clone https://github.com/rdashevsky/go-pkgs.git
cd go-pkgs

# Setup development environment
make dev-setup

# Run tests
make test

# Run all checks
make ci
```

### Available Commands
```bash
make help          # Show all available commands
make test          # Run tests
make lint          # Run linter
make build         # Build packages
make ci            # Run all CI checks
make security-check # Run security scans
```

## License
MIT License
