package server_test

import (
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rdashevsky/go-pkgs/rabbitmq/server"
)

func BenchmarkNew(b *testing.B) {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"test": func(d *amqp.Delivery) (interface{}, error) {
			return "response", nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			router,
			logger,
			server.ConnWaitTime(time.Millisecond),
			server.ConnAttempts(1),
		)
	}
}

func BenchmarkNewWithOptions(b *testing.B) {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"handler1": func(d *amqp.Delivery) (interface{}, error) {
			return map[string]string{"result": "ok"}, nil
		},
		"handler2": func(d *amqp.Delivery) (interface{}, error) {
			return []int{1, 2, 3}, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			router,
			logger,
			server.Timeout(time.Second),
			server.ConnWaitTime(time.Millisecond),
			server.ConnAttempts(1),
		)
	}
}

func BenchmarkServer_Shutdown(b *testing.B) {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"test": func(d *amqp.Delivery) (interface{}, error) {
			return "ok", nil
		},
	}

	servers := make([]*server.Server, b.N)
	for i := 0; i < b.N; i++ {
		s, err := server.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			router,
			logger,
			server.ConnWaitTime(time.Millisecond),
			server.ConnAttempts(1),
		)
		if err != nil {
			b.Skip("RabbitMQ server not available for benchmark")
		}
		servers[i] = s
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = servers[i].Shutdown()
	}
}

func BenchmarkServer_Notify(b *testing.B) {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"test": func(d *amqp.Delivery) (interface{}, error) {
			return "ok", nil
		},
	}

	s, err := server.New(
		"amqp://guest:guest@localhost:5672/",
		"benchmark-server",
		router,
		logger,
		server.ConnWaitTime(time.Millisecond),
		server.ConnAttempts(1),
	)
	if err != nil {
		b.Skip("RabbitMQ server not available for benchmark")
	}
	defer s.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Notify()
	}
}

func BenchmarkCallHandler_Simple(b *testing.B) {
	handler := func(d *amqp.Delivery) (interface{}, error) {
		return "simple response", nil
	}

	delivery := &amqp.Delivery{
		Body: []byte(`{"test": "data"}`),
		Type: "test-handler",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler(delivery)
	}
}

func BenchmarkCallHandler_Complex(b *testing.B) {
	handler := func(d *amqp.Delivery) (interface{}, error) {
		return map[string]interface{}{
			"status":    "success",
			"timestamp": time.Now().Unix(),
			"data":      string(d.Body),
			"type":      d.Type,
		}, nil
	}

	delivery := &amqp.Delivery{
		Body: []byte(`{"operation": "benchmark", "data": [1,2,3,4,5]}`),
		Type: "complex-handler",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler(delivery)
	}
}

func BenchmarkRouterLookup(b *testing.B) {
	router := map[string]server.CallHandler{
		"handler1": func(d *amqp.Delivery) (interface{}, error) {
			return "response1", nil
		},
		"handler2": func(d *amqp.Delivery) (interface{}, error) {
			return "response2", nil
		},
		"handler3": func(d *amqp.Delivery) (interface{}, error) {
			return "response3", nil
		},
		"handler4": func(d *amqp.Delivery) (interface{}, error) {
			return "response4", nil
		},
		"handler5": func(d *amqp.Delivery) (interface{}, error) {
			return "response5", nil
		},
	}

	handlerNames := []string{"handler1", "handler2", "handler3", "handler4", "handler5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := handlerNames[i%len(handlerNames)]
		_, exists := router[name]
		if !exists {
			b.Errorf("handler %s not found", name)
		}
	}
}

func BenchmarkMockLogger_Error(b *testing.B) {
	logger := &mockLogger{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("benchmark message")
	}
}

func BenchmarkServerOptions(b *testing.B) {
	logger := &mockLogger{}
	router := map[string]server.CallHandler{
		"test": func(d *amqp.Delivery) (interface{}, error) {
			return "ok", nil
		},
	}

	options := [][]server.Option{
		{server.Timeout(time.Second)},
		{server.ConnWaitTime(500 * time.Millisecond)},
		{server.ConnAttempts(3)},
		{
			server.Timeout(2 * time.Second),
			server.ConnWaitTime(time.Second),
			server.ConnAttempts(5),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := options[i%len(options)]
		opts = append(opts, server.ConnWaitTime(time.Millisecond), server.ConnAttempts(1))
		_, _ = server.New(
			"amqp://guest:guest@localhost:5672/",
			"benchmark-server",
			router,
			logger,
			opts...,
		)
	}
}

func BenchmarkServerCreationWithDifferentRouterSizes(b *testing.B) {
	logger := &mockLogger{}

	routerSizes := []int{1, 5, 10, 20, 50}

	for _, size := range routerSizes {
		b.Run(fmt.Sprintf("router-size-%d", size), func(b *testing.B) {
			router := make(map[string]server.CallHandler)
			for i := 0; i < size; i++ {
				handlerName := fmt.Sprintf("handler%d", i)
				router[handlerName] = func(d *amqp.Delivery) (interface{}, error) {
					return fmt.Sprintf("response%d", i), nil
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = server.New(
					"amqp://guest:guest@localhost:5672/",
					"benchmark-server",
					router,
					logger,
					server.ConnWaitTime(time.Millisecond),
					server.ConnAttempts(1),
				)
			}
		})
	}
}
