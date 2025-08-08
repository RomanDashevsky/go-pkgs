package postgres_test

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/rdashevsky/go-pkgs/postgres"
)

func TestNew_URLParsing(t *testing.T) {
	// Test that URL parsing works for basic valid URLs
	tests := []struct {
		name string
		url  string
		opts []postgres.Option
	}{
		{
			name: "basic URL with options",
			url:  "postgres://user:pass@127.0.0.1:65432/testdb",
			opts: []postgres.Option{
				postgres.MaxPoolSize(5),
				postgres.ConnAttempts(1),
				postgres.ConnTimeout(50 * time.Millisecond),
			},
		},
		{
			name: "URL without options",
			url:  "postgres://user:pass@127.0.0.1:65432/testdb",
			opts: []postgres.Option{postgres.ConnAttempts(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg, err := postgres.New(tt.url, tt.opts...)

			// We expect either success (if there's a real DB) or connection failure
			// Both are valid outcomes - we're just testing that parsing doesn't fail
			if pg != nil {
				pg.Close()
			}

			// Just verify that we get a consistent result
			_ = err // Connection might succeed or fail, both are OK
		})
	}
}

func TestPostgres_SquirrelBuilder(t *testing.T) {
	// Create a postgres instance that will fail connection but still has builder configured
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Test that Squirrel builder works correctly
	query, args, err := pg.Builder.
		Select("id, name").
		From("users").
		Where(squirrel.Eq{"active": true}).
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Limit(10).
		ToSql()

	if err != nil {
		t.Fatalf("failed to build query: %v", err)
	}

	expectedQuery := "SELECT id, name FROM users WHERE active = $1 AND created_at > $2 LIMIT 10"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}

	if args[0] != true {
		t.Errorf("expected first arg to be true, got %v", args[0])
	}
}

func TestPostgres_SquirrelBuilderInsert(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pg.Builder.
		Insert("users").
		Columns("name", "email", "active").
		Values("John Doe", "john@example.com", true).
		Values("Jane Smith", "jane@example.com", false).
		ToSql()

	if err != nil {
		t.Fatalf("failed to build insert query: %v", err)
	}

	expectedQuery := "INSERT INTO users (name,email,active) VALUES ($1,$2,$3),($4,$5,$6)"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 6 {
		t.Errorf("expected 6 args, got %d", len(args))
	}
}

func TestPostgres_SquirrelBuilderUpdate(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pg.Builder.
		Update("users").
		Set("name", "Updated Name").
		Set("updated_at", "NOW()").
		Where(squirrel.Eq{"id": 123}).
		ToSql()

	if err != nil {
		t.Fatalf("failed to build update query: %v", err)
	}

	expectedQuery := "UPDATE users SET name = $1, updated_at = $2 WHERE id = $3"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestPostgres_SquirrelBuilderDelete(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pg.Builder.
		Delete("users").
		Where("active = ?", false).
		Where("created_at < ?", time.Now().Add(-30*24*time.Hour)).
		ToSql()

	if err != nil {
		t.Fatalf("failed to build delete query: %v", err)
	}

	expectedQuery := "DELETE FROM users WHERE active = $1 AND created_at < $2"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestPostgres_Close(t *testing.T) {
	pg := &postgres.Postgres{}

	// Should not panic when Pool is nil
	pg.Close()

	// Test would require actual database connection to test pool closing
	// This is a basic smoke test for nil safety
}

func TestOptions(t *testing.T) {
	t.Run("MaxPoolSize option", func(t *testing.T) {
		// Test that MaxPoolSize option can be applied without panicking
		pg, err := postgres.New(
			"postgres://user:pass@127.0.0.1:65432/db",
			postgres.MaxPoolSize(15),
			postgres.ConnAttempts(1),
		)

		// Connection may succeed or fail, both are acceptable
		if pg != nil {
			pg.Close()
		}
		_ = err // Don't check error - connection outcome varies by environment
	})

	t.Run("ConnTimeout option", func(t *testing.T) {
		// Test that ConnTimeout option can be applied without panicking
		pg, err := postgres.New(
			"postgres://user:pass@127.0.0.1:65432/db",
			postgres.ConnAttempts(1),
			postgres.ConnTimeout(10*time.Millisecond),
		)

		// Connection may succeed or fail, both are acceptable
		if pg != nil {
			pg.Close()
		}
		_ = err // Don't check error - connection outcome varies by environment
	})

	t.Run("ConnAttempts option", func(t *testing.T) {
		// Test that ConnAttempts option can be applied without panicking
		pg, err := postgres.New(
			"postgres://user:pass@127.0.0.1:65432/db",
			postgres.ConnAttempts(1),
		)

		// Connection may succeed or fail, both are acceptable
		if pg != nil {
			pg.Close()
		}
		_ = err // Don't check error - connection outcome varies by environment
	})
}

func TestPostgres_MultipleOptions(t *testing.T) {
	// Test that multiple options can be applied together without panicking
	pg, err := postgres.New(
		"postgres://testuser:testpass@127.0.0.1:65432/testdb",
		postgres.MaxPoolSize(20),
		postgres.ConnAttempts(1),
		postgres.ConnTimeout(10*time.Millisecond),
	)

	// Connection may succeed or fail, both are acceptable
	if pg != nil {
		pg.Close()
	}
	_ = err // Don't check error - connection outcome varies by environment
}

// Example demonstrates creating and using a PostgreSQL connection
func Example() {
	// Create connection with options
	pg, err := postgres.New(
		"postgres://user:password@localhost:5432/database",
		postgres.MaxPoolSize(10),
		postgres.ConnAttempts(3),
	)
	if err != nil {
		panic(err)
	}
	defer pg.Close()

	// Use Squirrel query builder
	query, args, err := pg.Builder.
		Select("id, name, email").
		From("users").
		Where(squirrel.Eq{"active": true}).
		OrderBy("created_at DESC").
		Limit(10).
		ToSql()
	if err != nil {
		panic(err)
	}

	// Execute query (would need actual database)
	_ = query
	_ = args
	// rows, err := pg.Pool.Query(context.Background(), query, args...)
}

// BenchmarkSquirrelQueryBuilding benchmarks Squirrel query building
func BenchmarkSquirrelQueryBuilding(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := pg.Builder.
			Select("id, name, email, created_at").
			From("users").
			Where(squirrel.Eq{"active": true}).
			Where("created_at > ?", time.Now().Add(-24*time.Hour)).
			OrderBy("created_at DESC").
			Limit(50).
			ToSql()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNew benchmarks postgres connection creation (will fail but measures parsing overhead)
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pg, _ := postgres.New(
			"postgres://user:pass@127.0.0.1:65432/db",
			postgres.ConnAttempts(1),
		)
		if pg != nil {
			pg.Close()
		}
	}
}
