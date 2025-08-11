package goose_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rdashevsky/go-pkgs/goose"
	"github.com/rdashevsky/go-pkgs/logger"
)

// BenchmarkCheckMigrationStatus_ConnectionFailure benchmarks migration status check with connection failures
func BenchmarkCheckMigrationStatus_ConnectionFailure(b *testing.B) {
	config, err := pgxpool.ParseConfig("postgres://bench:bench@127.0.0.1:65432/nonexistent")
	if err != nil {
		b.Fatalf("failed to parse config: %v", err)
	}

	// Configure for quick failure to minimize benchmark time
	config.MaxConns = 1
	config.MinConns = 0
	config.ConnConfig.ConnectTimeout = 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		b.Skip("failed to create pool - expected in benchmark environment")
	}
	defer pool.Close()

	logger := &benchmarkLogger{}
	expectedVersion := int64(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = goose.CheckMigrationStatus(pool, expectedVersion, logger)
	}
}

// BenchmarkCheckMigrationStatus_DifferentVersions benchmarks with different expected versions
func BenchmarkCheckMigrationStatus_DifferentVersions(b *testing.B) {
	versions := []int64{0, 1, 10, 100, 1000, 999999}

	for _, version := range versions {
		b.Run(fmt.Sprintf("version_%d", version), func(b *testing.B) {
			config, err := pgxpool.ParseConfig("postgres://bench:bench@127.0.0.1:65432/benchdb")
			if err != nil {
				b.Fatalf("failed to parse config: %v", err)
			}

			config.MaxConns = 1
			config.ConnConfig.ConnectTimeout = 10 * time.Millisecond

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				b.Skip("failed to create pool - expected in benchmark environment")
			}
			defer pool.Close()

			logger := &benchmarkLogger{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = goose.CheckMigrationStatus(pool, version, logger)
			}
		})
	}
}

// BenchmarkCheckMigrationStatus_PoolConfigurations benchmarks with different pool configurations
func BenchmarkCheckMigrationStatus_PoolConfigurations(b *testing.B) {
	poolConfigs := []struct {
		name     string
		maxConns int32
		minConns int32
	}{
		{"small_pool", 1, 0},
		{"medium_pool", 5, 1},
		{"large_pool", 20, 5},
	}

	for _, pc := range poolConfigs {
		b.Run(pc.name, func(b *testing.B) {
			config, err := pgxpool.ParseConfig("postgres://bench:bench@127.0.0.1:65432/benchdb")
			if err != nil {
				b.Fatalf("failed to parse config: %v", err)
			}

			config.MaxConns = pc.maxConns
			config.MinConns = pc.minConns
			config.ConnConfig.ConnectTimeout = 10 * time.Millisecond

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				b.Skip("failed to create pool - expected in benchmark environment")
			}
			defer pool.Close()

			logger := &benchmarkLogger{}
			expectedVersion := int64(1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = goose.CheckMigrationStatus(pool, expectedVersion, logger)
			}
		})
	}
}

// BenchmarkCheckMigrationStatus_LoggerTypes benchmarks with different logger implementations
func BenchmarkCheckMigrationStatus_LoggerTypes(b *testing.B) {
	// Create different logger implementations
	loggerTypes := map[string]func() logger.LoggerI{
		"no_op": func() logger.LoggerI { return &benchmarkLogger{} },
		"real":  func() logger.LoggerI { return logger.New("error") }, // Use error level to minimize output
		"mock":  func() logger.LoggerI { return &mockLogger{} },
	}

	for name, createLogger := range loggerTypes {
		b.Run(name, func(b *testing.B) {
			config, err := pgxpool.ParseConfig("postgres://bench:bench@127.0.0.1:65432/benchdb")
			if err != nil {
				b.Fatalf("failed to parse config: %v", err)
			}

			config.MaxConns = 1
			config.ConnConfig.ConnectTimeout = 10 * time.Millisecond

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				b.Skip("failed to create pool - expected in benchmark environment")
			}
			defer pool.Close()

			logger := createLogger()
			expectedVersion := int64(1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = goose.CheckMigrationStatus(pool, expectedVersion, logger)
			}
		})
	}
}

// BenchmarkCheckMigrationStatus_Concurrent benchmarks concurrent access
func BenchmarkCheckMigrationStatus_Concurrent(b *testing.B) {
	config, err := pgxpool.ParseConfig("postgres://bench:bench@127.0.0.1:65432/benchdb")
	if err != nil {
		b.Fatalf("failed to parse config: %v", err)
	}

	config.MaxConns = 10
	config.ConnConfig.ConnectTimeout = 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		b.Skip("failed to create pool - expected in benchmark environment")
	}
	defer pool.Close()

	logger := &benchmarkLogger{}
	expectedVersion := int64(1)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = goose.CheckMigrationStatus(pool, expectedVersion, logger)
		}
	})
}

// BenchmarkMockLogger_Operations benchmarks mock logger operations
func BenchmarkMockLogger_Operations(b *testing.B) {
	operations := map[string]func(*mockLogger){
		"debug": func(m *mockLogger) { m.Debug("debug message") },
		"info":  func(m *mockLogger) { m.Info("info message %d", 42) },
		"warn":  func(m *mockLogger) { m.Warn("warn message") },
		"error": func(m *mockLogger) { m.Error("error message") },
		"fatal": func(m *mockLogger) { m.Fatal("fatal message") },
	}

	for name, op := range operations {
		b.Run(name, func(b *testing.B) {
			logger := &mockLogger{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				op(logger)
			}
		})
	}
}

// BenchmarkMockLogger_GetLogsByLevel benchmarks log retrieval by level
func BenchmarkMockLogger_GetLogsByLevel(b *testing.B) {
	logger := &mockLogger{}

	// Pre-populate with logs
	for i := 0; i < 1000; i++ {
		logger.Debug(fmt.Sprintf("debug %d", i))
		logger.Info("info %d", i)
		logger.Error(fmt.Sprintf("error %d", i))
	}

	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	for _, level := range levels {
		b.Run(level, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = logger.getLogsByLevel(level)
			}
		})
	}
}

// BenchmarkPgxPoolConfig_Creation benchmarks pool configuration parsing and creation
func BenchmarkPgxPoolConfig_Creation(b *testing.B) {
	dsns := []string{
		"postgres://user:pass@localhost:5432/db",
		"postgres://user:pass@127.0.0.1:5432/db?sslmode=disable",
		"postgres://user:pass@host:5432/db?sslmode=require&pool_max_conns=10",
	}

	for i, dsn := range dsns {
		b.Run(fmt.Sprintf("dsn_%d", i), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				config, err := pgxpool.ParseConfig(dsn)
				if err != nil {
					b.Fatal(err)
				}

				// Simulate minimal pool creation setup
				config.MaxConns = 1
				config.ConnConfig.ConnectTimeout = 1 * time.Millisecond

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()

				// Note: This will fail to connect, but we're benchmarking config creation
				pool, err := pgxpool.NewWithConfig(ctx, config)
				if err == nil {
					pool.Close()
				}
			}
		})
	}
}
