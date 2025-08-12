package rabbitmq_test

import (
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq"
)

func BenchmarkNew(b *testing.B) {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: time.Second,
		Attempts: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rabbitmq.New("benchmark-exchange", cfg)
	}
}

func BenchmarkNewWithDifferentConfigs(b *testing.B) {
	configs := []rabbitmq.Config{
		{
			URL:      "amqp://guest:guest@localhost:5672/",
			WaitTime: time.Second,
			Attempts: 1,
		},
		{
			URL:      "amqp://guest:guest@localhost:5672/",
			WaitTime: 5 * time.Second,
			Attempts: 3,
		},
		{
			URL:      "amqp://user:pass@localhost:5672/vhost",
			WaitTime: 10 * time.Second,
			Attempts: 5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := configs[i%len(configs)]
		_ = rabbitmq.New("benchmark-exchange", cfg)
	}
}

func BenchmarkConnection_AttemptConnect(b *testing.B) {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: 10 * time.Millisecond,
		Attempts: 1,
	}

	conns := make([]*rabbitmq.Connection, b.N)
	for i := 0; i < b.N; i++ {
		conns[i] = rabbitmq.New("benchmark-exchange", cfg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conns[i].AttemptConnect()
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		// Clean up immediately to avoid resource exhaustion
		if conns[i].Connection != nil {
			conns[i].Connection.Close()
		}
	}
}

func BenchmarkConnection_AttemptConnectWithRetry(b *testing.B) {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: time.Millisecond,
		Attempts: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := rabbitmq.New("benchmark-retry", cfg)
		err := conn.AttemptConnect()
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		// Clean up
		if conn.Connection != nil {
			conn.Connection.Close()
		}
	}
}

func BenchmarkConnection_AttemptConnectFailure(b *testing.B) {
	cfg := rabbitmq.Config{
		URL:      "amqp://nonexistent-host:5672/",
		WaitTime: time.Microsecond, // Very short wait time for benchmark
		Attempts: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := rabbitmq.New("benchmark-fail", cfg)
		_ = conn.AttemptConnect() // Expected to fail
	}
}

func BenchmarkConnection_ConfigVariations(b *testing.B) {
	waitTimes := []time.Duration{
		time.Microsecond,
		time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
	}

	attempts := []int{1, 2, 3, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := rabbitmq.Config{
			URL:      "amqp://guest:guest@localhost:5672/",
			WaitTime: waitTimes[i%len(waitTimes)],
			Attempts: attempts[i%len(attempts)],
		}
		conn := rabbitmq.New("benchmark-variations", cfg)
		err := conn.AttemptConnect()
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		if conn.Connection != nil {
			conn.Connection.Close()
		}
	}
}

func BenchmarkConnection_ExchangeNameLength(b *testing.B) {
	cfg := rabbitmq.Config{
		URL:      "amqp://guest:guest@localhost:5672/",
		WaitTime: 10 * time.Millisecond,
		Attempts: 1,
	}

	exchangeNames := []string{
		"a",
		"short-exchange",
		"medium-length-exchange-name",
		"very-long-exchange-name-that-is-quite-detailed-and-descriptive",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exchangeName := exchangeNames[i%len(exchangeNames)]
		conn := rabbitmq.New(exchangeName, cfg)
		err := conn.AttemptConnect()
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		if conn.Connection != nil {
			conn.Connection.Close()
		}
	}
}
