package client

import (
	"testing"
	"time"

	kafka "github.com/rdashevsky/go-pkgs/kafka"
)

func TestCallTimeout(t *testing.T) {
	timeout := 15 * time.Second
	cfg := kafka.Config{
		Brokers:    []string{"localhost:9092"},
		ClientID:   "test-client",
		GroupID:    "test-group",
		AutoCommit: true,
	}

	client, err := New(cfg, "test-requests", "test-replies", CallTimeout(timeout))
	if err != nil {
		// Skip test if Kafka is not available
		t.Skipf("Skipping test - Kafka not available: %v", err)
	}
	defer func() {
		if client != nil {
			_ = client.Shutdown()
		}
	}()

	if client.callTimeout != timeout {
		t.Errorf("Expected call timeout %v, got %v", timeout, client.callTimeout)
	}
}
