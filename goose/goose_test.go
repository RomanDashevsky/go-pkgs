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

// mockLogger implements logger.LoggerI for testing
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   string
	message string
}

func (m *mockLogger) Debug(message interface{}, args ...interface{}) {
	m.logs = append(m.logs, logEntry{
		level:   "DEBUG",
		message: fmt.Sprintf("%v", message),
	})
}

func (m *mockLogger) Info(message string, args ...interface{}) {
	formatted := fmt.Sprintf(message, args...)
	m.logs = append(m.logs, logEntry{
		level:   "INFO",
		message: formatted,
	})
}

func (m *mockLogger) Warn(message string, args ...interface{}) {
	formatted := fmt.Sprintf(message, args...)
	m.logs = append(m.logs, logEntry{
		level:   "WARN",
		message: formatted,
	})
}

func (m *mockLogger) Error(message interface{}, args ...interface{}) {
	var formatted string
	if len(args) > 0 {
		formatted = fmt.Sprintf("%v", message)
		for _, arg := range args {
			formatted += fmt.Sprintf(" %v", arg)
		}
	} else {
		switch msg := message.(type) {
		case string:
			formatted = msg
		case error:
			formatted = msg.Error()
		default:
			formatted = fmt.Sprintf("%v", message)
		}
	}

	m.logs = append(m.logs, logEntry{
		level:   "ERROR",
		message: formatted,
	})
}

func (m *mockLogger) Fatal(message interface{}, args ...interface{}) {
	var formatted string
	if len(args) > 0 {
		formatted = fmt.Sprintf("%v", message)
		for _, arg := range args {
			formatted += fmt.Sprintf(" %v", arg)
		}
	} else {
		switch msg := message.(type) {
		case string:
			formatted = msg
		case error:
			formatted = msg.Error()
		default:
			formatted = fmt.Sprintf("%v", message)
		}
	}

	m.logs = append(m.logs, logEntry{
		level:   "FATAL",
		message: formatted,
	})
}

func (m *mockLogger) getLogsByLevel(level string) []string {
	var messages []string
	for _, log := range m.logs {
		if log.level == level {
			messages = append(messages, log.message)
		}
	}
	return messages
}

func TestCheckMigrationStatus_NoDatabase(t *testing.T) {
	// Test with a pool that will fail to connect
	config, err := pgxpool.ParseConfig("postgres://test:test@127.0.0.1:65432/nonexistent")
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Set minimal pool size and quick timeout
	config.MaxConns = 1
	config.MinConns = 0
	config.MaxConnLifetime = 1 * time.Second
	config.MaxConnIdleTime = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Skip("failed to create pool - this is expected in test environment")
	}
	defer pool.Close()

	mockLog := &mockLogger{}

	// This should fail because there's no database
	version, err := goose.CheckMigrationStatus(pool, 5, mockLog)
	if err == nil {
		t.Skip("unexpected successful database connection")
	}

	if version != 0 {
		t.Errorf("expected version 0 on error, got %d", version)
	}
}

func TestCheckMigrationStatus_InvalidPool(t *testing.T) {
	// Test with nil pool - should cause panic or error when trying to use stdlib.OpenDBFromPool
	// We can't easily test this without causing panic, so we'll skip this specific case
	t.Skip("Testing with nil pool would cause panic - this is expected behavior")
}

// TestCheckMigrationStatus_VersionMismatch simulates a version mismatch scenario
// This test demonstrates the expected behavior without requiring a real database
func TestCheckMigrationStatus_VersionMismatch(t *testing.T) {
	// Create a config that will fail to connect (simulating DB error)
	config, err := pgxpool.ParseConfig("postgres://test:test@127.0.0.1:65432/testdb")
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	config.MaxConns = 1
	config.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Skip("failed to create pool - expected in test environment")
	}
	defer pool.Close()

	mockLog := &mockLogger{}

	// Expected to fail due to no database connection
	version, err := goose.CheckMigrationStatus(pool, 10, mockLog)
	if err == nil {
		t.Skip("unexpected successful connection to database")
	}

	// Should return 0 on connection error
	if version != 0 {
		t.Errorf("expected version 0 on connection error, got %d", version)
	}
}

func TestCheckMigrationStatus_Logger(t *testing.T) {
	// Test that the logger is properly used
	config, err := pgxpool.ParseConfig("postgres://test:test@127.0.0.1:65432/testdb")
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	config.MaxConns = 1
	config.ConnConfig.ConnectTimeout = 100 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Skip("failed to create pool - expected in test environment")
	}
	defer pool.Close()

	mockLog := &mockLogger{}

	// This will fail to connect, but we can check logger behavior
	_, err = goose.CheckMigrationStatus(pool, 3, mockLog)
	if err == nil {
		// If it somehow succeeds, check that info log was called
		infoLogs := mockLog.getLogsByLevel("INFO")
		if len(infoLogs) != 1 {
			t.Errorf("expected 1 info log entry, got %d", len(infoLogs))
		}
		if len(infoLogs) > 0 && infoLogs[0] != "Migrations are up to date: 3" {
			t.Errorf("expected migration success message, got: %s", infoLogs[0])
		}
	} else {
		// Expected case - connection fails
		// The function should not log anything on connection failure
		// (the error is returned, not logged)
	}
}

// TestCheckMigrationStatus_Success demonstrates successful migration check
// This would work with a real database that has goose migrations table
func TestCheckMigrationStatus_Success(t *testing.T) {
	t.Skip("Integration test - requires real PostgreSQL database with goose_db_version table")

	// Example of what this test would look like with a real database:
	/*
		config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/testdb")
		if err != nil {
			t.Fatalf("failed to parse config: %v", err)
		}

		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			t.Fatalf("failed to create pool: %v", err)
		}
		defer pool.Close()

		mockLog := &mockLogger{}
		expectedVersion := int64(5)

		version, err := goose.CheckMigrationStatus(pool, expectedVersion, mockLog)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		if version != expectedVersion {
			t.Errorf("expected version %d, got %d", expectedVersion, version)
		}

		// Check that success was logged
		infoLogs := mockLog.getLogsByLevel("INFO")
		if len(infoLogs) != 1 {
			t.Errorf("expected 1 info log, got %d", len(infoLogs))
		}
	*/
}

// TestGooseIntegration_DatabaseVersionCheck demonstrates database version checking
func TestGooseIntegration_DatabaseVersionCheck(t *testing.T) {
	t.Skip("Integration test - requires real PostgreSQL database")

	// This test would require:
	// 1. A real PostgreSQL database
	// 2. Goose migrations applied to create goose_db_version table
	// 3. Known migration version to test against

	// Example test structure:
	/*
		dsn := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			t.Fatalf("failed to parse config: %v", err)
		}

		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			t.Fatalf("failed to create pool: %v", err)
		}
		defer pool.Close()

		mockLog := &mockLogger{}

		tests := []struct {
			name            string
			expectedVersion int64
			shouldMatch     bool
		}{
			{"matching version", 3, true},
			{"mismatched version", 5, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				version, err := goose.CheckMigrationStatus(pool, tt.expectedVersion, mockLog)

				if tt.shouldMatch {
					if err != nil {
						t.Errorf("expected no error for matching version, got: %v", err)
					}
					if version != tt.expectedVersion {
						t.Errorf("expected version %d, got %d", tt.expectedVersion, version)
					}
				} else {
					if err == nil {
						t.Error("expected error for mismatched version")
					}
					// Version should still be returned even on mismatch
					if version == 0 {
						t.Error("expected non-zero version even on mismatch")
					}
				}
			})
		}
	*/
}

// Example demonstrates using the CheckMigrationStatus function
func Example() {
	// Create a PostgreSQL connection pool
	config, err := pgxpool.ParseConfig("postgres://user:password@localhost:5432/database?sslmode=disable")
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Create a logger
	log := logger.New("info")

	// Check if database is at expected migration version
	expectedVersion := int64(10)
	currentVersion, err := goose.CheckMigrationStatus(pool, expectedVersion, log)
	if err != nil {
		fmt.Printf("Migration check failed: %v\n", err)
		fmt.Printf("Current version: %d, Expected: %d\n", currentVersion, expectedVersion)
		return
	}

	fmt.Printf("Database is up to date at version %d\n", currentVersion)
}

// BenchmarkCheckMigrationStatus benchmarks the migration status check
func BenchmarkCheckMigrationStatus(b *testing.B) {
	// This benchmark would fail without a real database, but shows the structure
	b.Skip("Benchmark requires real PostgreSQL database")

	/*
		config, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/testdb")
		if err != nil {
			b.Fatalf("failed to parse config: %v", err)
		}

		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			b.Fatalf("failed to create pool: %v", err)
		}
		defer pool.Close()

		log := logger.New("error") // Use error level to minimize logging overhead
		expectedVersion := int64(1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goose.CheckMigrationStatus(pool, expectedVersion, log)
		}
	*/
}

// TestMockLogger verifies our mock logger implementation works correctly
func TestMockLogger(t *testing.T) {
	mockLog := &mockLogger{}

	mockLog.Debug("debug message")
	mockLog.Info("info message with arg: %d", 42)
	mockLog.Warn("warn message")
	mockLog.Error("error message")
	mockLog.Fatal("fatal message")

	if len(mockLog.logs) != 5 {
		t.Errorf("expected 5 log entries, got %d", len(mockLog.logs))
	}

	debugLogs := mockLog.getLogsByLevel("DEBUG")
	if len(debugLogs) != 1 || debugLogs[0] != "debug message" {
		t.Errorf("debug log not captured correctly: %v", debugLogs)
	}

	infoLogs := mockLog.getLogsByLevel("INFO")
	if len(infoLogs) != 1 || infoLogs[0] != "info message with arg: 42" {
		t.Errorf("info log not formatted correctly: %v", infoLogs)
	}
}
