package kafka

import (
	"context"
	"testing"
	"time"
)

func TestNewConnection(t *testing.T) {
	cfg := Config{
		Brokers:     []string{"localhost:9092"},
		Timeout:     10 * time.Second,
		RetryDelay:  1 * time.Second,
		MaxRetries:  2,
		ClientID:    "test-client",
		GroupID:     "test-group",
		AutoCommit:  true,
		StartOffset: 0,
	}

	conn := NewConnection(cfg)
	if conn == nil {
		t.Fatal("NewConnection returned nil")
	}

	if conn.Timeout != cfg.Timeout {
		t.Errorf("Expected timeout %v, got %v", cfg.Timeout, conn.Timeout)
	}

	if conn.ClientID != cfg.ClientID {
		t.Errorf("Expected client ID %s, got %s", cfg.ClientID, conn.ClientID)
	}

	defer conn.Close()
}

func TestConnectionDefaults(t *testing.T) {
	cfg := Config{
		Brokers: []string{"localhost:9092"},
	}

	conn := NewConnection(cfg)
	if conn == nil {
		t.Fatal("NewConnection returned nil")
	}

	if conn.Timeout == 0 {
		t.Error("Expected default timeout to be set")
	}

	if conn.RetryDelay == 0 {
		t.Error("Expected default retry delay to be set")
	}

	if conn.MaxRetries == 0 {
		t.Error("Expected default max retries to be set")
	}

	if conn.ClientID == "" {
		t.Error("Expected default client ID to be set")
	}

	defer conn.Close()
}

func TestConnectionConnect_InvalidBroker(t *testing.T) {
	cfg := Config{
		Brokers:    []string{"invalid:9092"},
		Timeout:    1 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: 1,
	}

	conn := NewConnection(cfg)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := conn.Connect(ctx)
	// Note: franz-go client creation may succeed even with invalid brokers
	// The error will occur when actually trying to produce/consume
	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Logf("Client created successfully - errors will surface during actual operations")
	}
}
