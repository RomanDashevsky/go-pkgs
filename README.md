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
- Go 1.21 or later
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
