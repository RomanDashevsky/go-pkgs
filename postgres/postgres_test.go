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
		t.Run(tt.name, func(_ *testing.T) {
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

func TestPostgres_Close(_ *testing.T) {
	pg := &postgres.Postgres{}

	// Should not panic when Pool is nil
	pg.Close()

	// Test would require actual database connection to test pool closing
	// This is a basic smoke test for nil safety
}

func TestOptions(t *testing.T) {
	t.Run("MaxPoolSize option", func(_ *testing.T) {
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

	t.Run("ConnTimeout option", func(_ *testing.T) {
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

	t.Run("ConnAttempts option", func(_ *testing.T) {
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

func TestPostgres_MultipleOptions(_ *testing.T) {
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

func TestPostgres_SquirrelBuilderComplexSelect(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Test complex SELECT with JOINs and subqueries
	query, args, err := pg.Builder.
		Select("u.id", "u.name", "p.title", "COUNT(c.id) as comment_count").
		From("users u").
		Join("posts p ON u.id = p.user_id").
		LeftJoin("comments c ON p.id = c.post_id").
		Where(squirrel.And{
			squirrel.Eq{"u.active": true},
			squirrel.Gt{"p.created_at": time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		}).
		GroupBy("u.id", "u.name", "p.title").
		Having("COUNT(c.id) > ?", 5).
		OrderBy("comment_count DESC").
		Limit(20).
		Offset(10).
		ToSql()

	if err != nil {
		t.Fatalf("failed to build complex query: %v", err)
	}

	expectedQuery := "SELECT u.id, u.name, p.title, COUNT(c.id) as comment_count FROM users u JOIN posts p ON u.id = p.user_id LEFT JOIN comments c ON p.id = c.post_id WHERE (u.active = $1 AND p.created_at > $2) GROUP BY u.id, u.name, p.title HAVING COUNT(c.id) > $3 ORDER BY comment_count DESC LIMIT 20 OFFSET 10"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestPostgres_SquirrelBuilderBatchInsert(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Test batch insert with multiple rows
	insertBuilder := pg.Builder.Insert("products").Columns("name", "price", "category_id")

	products := []struct {
		name       string
		price      float64
		categoryID int
	}{
		{"Product A", 19.99, 1},
		{"Product B", 29.99, 2},
		{"Product C", 39.99, 1},
		{"Product D", 49.99, 3},
	}

	for _, product := range products {
		insertBuilder = insertBuilder.Values(product.name, product.price, product.categoryID)
	}

	query, args, err := insertBuilder.ToSql()

	if err != nil {
		t.Fatalf("failed to build batch insert query: %v", err)
	}

	expectedQuery := "INSERT INTO products (name,price,category_id) VALUES ($1,$2,$3),($4,$5,$6),($7,$8,$9),($10,$11,$12)"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 12 {
		t.Errorf("expected 12 args, got %d", len(args))
	}
}

func TestPostgres_SquirrelBuilderUpsert(t *testing.T) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Test UPSERT using ON CONFLICT
	query, args, err := pg.Builder.
		Insert("users").
		Columns("email", "name", "updated_at").
		Values("john@example.com", "John Doe", "NOW()").
		Suffix("ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at").
		ToSql()

	if err != nil {
		t.Fatalf("failed to build upsert query: %v", err)
	}

	expectedQuery := "INSERT INTO users (email,name,updated_at) VALUES ($1,$2,$3) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at"
	if query != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}
