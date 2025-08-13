package server

import (
	"testing"

	kafka "github.com/rdashevsky/go-pkgs/kafka"
	"github.com/rdashevsky/go-pkgs/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestNew_MissingGroupID(t *testing.T) {
	cfg := kafka.Config{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-server",
		// GroupID is missing
	}

	router := map[string]CallHandler{
		"test": func(*kgo.Record) (interface{}, error) {
			return "ok", nil
		},
	}

	logger := logger.New("info")

	_, err := New(cfg, "test-topic", router, logger)
	if err == nil {
		t.Error("Expected error when GroupID is missing")
	}

	if err.Error() != "kafka_rpc server - NewServer - GroupID is required for server" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
