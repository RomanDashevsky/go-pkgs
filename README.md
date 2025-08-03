# Go Packages

Репозиторий с общими Go пакетами для переиспользования между микросервисами.

## Пакеты

### Logger
Логгер на основе zerolog с настраиваемыми уровнями.
```go
import "github.com/rdashevsky/go-pkgs/logger"

l := logger.New("info")
l.Info("Application started")
```

### HTTP Server  
HTTP сервер на базе Fiber с настраиваемыми опциями.
```go
import "github.com/rdashevsky/go-pkgs/httpserver"

server := httpserver.New(
    httpserver.Port(":8080"),
    httpserver.ReadTimeout(10 * time.Second),
)
```

### PostgreSQL
Клиент PostgreSQL с пулом соединений и Squirrel query builder.
```go
import "github.com/rdashevsky/go-pkgs/postgres"

pg, err := postgres.New(databaseURL,
    postgres.MaxPoolSize(10),
    postgres.ConnAttempts(5),
)
```

### Redis
Клиент Redis с настраиваемым TTL.
```go
import "github.com/rdashevsky/go-pkgs/redis"

r, err := redis.New("localhost:6379", "", "",
    redis.WithTTL(5 * time.Minute),
)
```

### Goose
Утилиты для работы с миграциями базы данных.
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

## Использование

1. Добавьте модуль в ваш `go.mod`:
```bash
go get github.com/rdashevsky/go-pkgs
```

2. Импортируйте нужные пакеты:
```go
import (
    "github.com/rdashevsky/go-pkgs/logger"
    "github.com/rdashevsky/go-pkgs/httpserver"
)
```

## Лицензия
MIT License
