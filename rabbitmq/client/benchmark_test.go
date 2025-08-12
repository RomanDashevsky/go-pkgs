package client_test

import (
	"testing"
	"time"

	"github.com/rdashevsky/go-pkgs/rabbitmq/client"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = client.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			"benchmark-client",
			client.ConnWaitTime(time.Millisecond),
			client.ConnAttempts(1),
		)
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = client.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			"benchmark-client",
			client.Timeout(time.Second),
			client.ConnWaitTime(time.Millisecond),
			client.ConnAttempts(1),
		)
	}
}

func BenchmarkClient_RemoteCall(b *testing.B) {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"benchmark-server",
		"benchmark-client",
		client.Timeout(100*time.Millisecond), // Short timeout for benchmark
	)
	if err != nil {
		b.Skip("RabbitMQ server not available for benchmark")
	}
	defer c.Shutdown()

	request := map[string]string{"benchmark": "data"}
	var response interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will likely timeout since no server is handling requests
		_ = c.RemoteCall("benchmark-handler", request, &response)
	}
}

func BenchmarkClient_Shutdown(b *testing.B) {
	clients := make([]*client.Client, b.N)
	for i := 0; i < b.N; i++ {
		c, err := client.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			"benchmark-client",
		)
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		clients[i] = c
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = clients[i].Shutdown()
	}
}

func BenchmarkClient_Notify(b *testing.B) {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"benchmark-server",
		"benchmark-client",
	)
	if err != nil {
		b.Skip("RabbitMQ server not available for benchmark")
	}
	defer c.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Notify()
	}
}

func BenchmarkMessage_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &client.Message{
			Queue:         "benchmark-queue",
			Priority:      1,
			ContentType:   "application/json",
			Body:          []byte(`{"benchmark": "test"}`),
			ReplyTo:       "reply-queue",
			CorrelationID: "corr-id",
		}
		_ = msg
	}
}

func BenchmarkClient_OptionApplication(b *testing.B) {
	options := [][]client.Option{
		{client.Timeout(time.Second)},
		{client.ConnWaitTime(500 * time.Millisecond)},
		{client.ConnAttempts(3)},
		{
			client.Timeout(2 * time.Second),
			client.ConnWaitTime(time.Second),
			client.ConnAttempts(5),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := options[i%len(options)]
		_, _ = client.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			"benchmark-client",
			opts...,
		)
	}
}

func BenchmarkClient_ConcurrentRemoteCalls(b *testing.B) {
	c, err := client.New(
		"amqp://guest:guest@localhost:5672/",
		"benchmark-server",
		"benchmark-client",
		client.Timeout(50*time.Millisecond),
	)
	if err != nil {
		b.Skip("RabbitMQ server not available for benchmark")
	}
	defer c.Shutdown()

	request := map[string]string{"concurrent": "test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var response interface{}
			_ = c.RemoteCall("concurrent-handler", request, &response)
		}
	})
}

func BenchmarkClient_DifferentTimeouts(b *testing.B) {
	timeouts := []time.Duration{
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
	}

	for _, timeout := range timeouts {
		b.Run("timeout-"+timeout.String(), func(b *testing.B) {
			c, err := client.New(
				"amqp://guest:guest@localhost:5672/",
				"benchmark-server",
				"benchmark-client",
				client.Timeout(timeout),
			)
			if err != nil {
				b.Skip("RabbitMQ server not available for benchmark")
			}
			defer c.Shutdown()

			request := map[string]string{"timeout": "test"}
			var response interface{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = c.RemoteCall("timeout-handler", request, &response)
			}
		})
	}
}

func BenchmarkClient_ConnectionVariations(b *testing.B) {
	variations := []struct {
		name string
		opts []client.Option
	}{
		{
			name: "default",
			opts: []client.Option{},
		},
		{
			name: "fast-connection",
			opts: []client.Option{
				client.ConnWaitTime(10 * time.Millisecond),
				client.ConnAttempts(1),
			},
		},
		{
			name: "slow-connection",
			opts: []client.Option{
				client.ConnWaitTime(time.Second),
				client.ConnAttempts(3),
			},
		},
		{
			name: "many-attempts",
			opts: []client.Option{
				client.ConnWaitTime(50 * time.Millisecond),
				client.ConnAttempts(10),
			},
		},
	}

	for _, variation := range variations {
		b.Run(variation.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = client.New(
					"amqp://guest:guest@localhost:5672/",
					"benchmark-server",
					"benchmark-client",
					variation.opts...,
				)
			}
		})
	}
}
