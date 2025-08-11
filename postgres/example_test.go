package postgres_test

import (
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/rdashevsky/go-pkgs/postgres"
)

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

// ExampleNew demonstrates different ways to create a PostgreSQL connection
func ExampleNew() {
	// Basic connection
	pg1, err := postgres.New("postgres://user:password@localhost:5432/database")
	if err != nil {
		log.Fatal(err)
	}
	defer pg1.Close()

	// Connection with custom pool size
	pg2, err := postgres.New(
		"postgres://user:password@localhost:5432/database",
		postgres.MaxPoolSize(20),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pg2.Close()

	// Connection with timeout and retry settings
	pg3, err := postgres.New(
		"postgres://user:password@localhost:5432/database",
		postgres.ConnTimeout(30*time.Second),
		postgres.ConnAttempts(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pg3.Close()

	// Connection with all options
	pg4, err := postgres.New(
		"postgres://user:password@localhost:5432/database",
		postgres.MaxPoolSize(15),
		postgres.ConnTimeout(10*time.Second),
		postgres.ConnAttempts(3),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pg4.Close()

	fmt.Println("All connections created successfully")
	// Output: All connections created successfully
}

// ExamplePostgres_Builder_select demonstrates building SELECT queries
func ExamplePostgres_Builder_select() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Simple SELECT
	query, args, err := pg.Builder.
		Select("id, name").
		From("users").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// SELECT with WHERE conditions
	query2, args2, err := pg.Builder.
		Select("id, name, email").
		From("users").
		Where(squirrel.Eq{"active": true}).
		Where("age > ?", 18).
		OrderBy("name").
		Limit(10).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: SELECT id, name FROM users
	// Args: []
	// Query: SELECT id, name, email FROM users WHERE active = $1 AND age > $2 ORDER BY name LIMIT 10
	// Args: [true 18]
}

// ExamplePostgres_Builder_insert demonstrates building INSERT queries
func ExamplePostgres_Builder_insert() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Single INSERT
	query, args, err := pg.Builder.
		Insert("users").
		Columns("name", "email", "active").
		Values("John Doe", "john@example.com", true).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// Batch INSERT
	insertBuilder := pg.Builder.Insert("products").Columns("name", "price")
	products := []struct {
		name  string
		price float64
	}{
		{"Product A", 19.99},
		{"Product B", 29.99},
		{"Product C", 39.99},
	}

	for _, product := range products {
		insertBuilder = insertBuilder.Values(product.name, product.price)
	}

	query2, args2, err := insertBuilder.ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: INSERT INTO users (name,email,active) VALUES ($1,$2,$3)
	// Args: [John Doe john@example.com true]
	// Query: INSERT INTO products (name,price) VALUES ($1,$2),($3,$4),($5,$6)
	// Args: [Product A 19.99 Product B 29.99 Product C 39.99]
}

// ExamplePostgres_Builder_update demonstrates building UPDATE queries
func ExamplePostgres_Builder_update() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Simple UPDATE
	query, args, err := pg.Builder.
		Update("users").
		Set("name", "Updated Name").
		Where(squirrel.Eq{"id": 123}).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// UPDATE with multiple fields and complex WHERE
	query2, args2, err := pg.Builder.
		Update("users").
		Set("name", "John Smith").
		Set("email", "john.smith@example.com").
		Set("updated_at", "NOW()").
		Where(squirrel.And{
			squirrel.Eq{"active": true},
			squirrel.Lt{"last_login": time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		}).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: UPDATE users SET name = $1 WHERE id = $2
	// Args: [Updated Name 123]
	// Query: UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE (active = $4 AND last_login < $5)
	// Args: [John Smith john.smith@example.com NOW() true 2023-01-01 00:00:00 +0000 UTC]
}

// ExamplePostgres_Builder_delete demonstrates building DELETE queries
func ExamplePostgres_Builder_delete() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Simple DELETE
	query, args, err := pg.Builder.
		Delete("users").
		Where(squirrel.Eq{"id": 123}).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// DELETE with multiple conditions
	fixedTime := time.Date(2025, 8, 11, 0, 0, 0, 0, time.UTC)
	query2, args2, err := pg.Builder.
		Delete("sessions").
		Where("expires_at < ?", fixedTime).
		Where(squirrel.Eq{"active": false}).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: DELETE FROM users WHERE id = $1
	// Args: [123]
	// Query: DELETE FROM sessions WHERE expires_at < $1 AND active = $2
	// Args: [2025-08-11 00:00:00 +0000 UTC false]
}

// ExamplePostgres_Builder_joins demonstrates building queries with JOINs
func ExamplePostgres_Builder_joins() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Query with INNER JOIN
	query, args, err := pg.Builder.
		Select("u.name", "p.title").
		From("users u").
		Join("posts p ON u.id = p.user_id").
		Where(squirrel.Eq{"u.active": true}).
		OrderBy("p.created_at DESC").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// Query with multiple JOINs and aggregation
	query2, args2, err := pg.Builder.
		Select("u.name", "COUNT(p.id) as post_count", "COUNT(c.id) as comment_count").
		From("users u").
		LeftJoin("posts p ON u.id = p.user_id").
		LeftJoin("comments c ON p.id = c.post_id").
		Where(squirrel.Eq{"u.active": true}).
		GroupBy("u.id", "u.name").
		Having("COUNT(p.id) > ?", 0).
		OrderBy("post_count DESC").
		Limit(10).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = $1 ORDER BY p.created_at DESC
	// Args: [true]
	// Query: SELECT u.name, COUNT(p.id) as post_count, COUNT(c.id) as comment_count FROM users u LEFT JOIN posts p ON u.id = p.user_id LEFT JOIN comments c ON p.id = c.post_id WHERE u.active = $1 GROUP BY u.id, u.name HAVING COUNT(p.id) > $2 ORDER BY post_count DESC LIMIT 10
	// Args: [true 0]
}

// ExamplePostgres_Builder_upsert demonstrates building UPSERT queries
func ExamplePostgres_Builder_upsert() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// UPSERT with ON CONFLICT
	query, args, err := pg.Builder.
		Insert("users").
		Columns("email", "name", "created_at").
		Values("john@example.com", "John Doe", "NOW()").
		Suffix("ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// UPSERT with DO NOTHING
	query2, args2, err := pg.Builder.
		Insert("user_preferences").
		Columns("user_id", "preference_key", "preference_value").
		Values(1, "theme", "dark").
		Values(1, "language", "en").
		Suffix("ON CONFLICT (user_id, preference_key) DO NOTHING").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: INSERT INTO users (email,name,created_at) VALUES ($1,$2,$3) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()
	// Args: [john@example.com John Doe NOW()]
	// Query: INSERT INTO user_preferences (user_id,preference_key,preference_value) VALUES ($1,$2,$3),($4,$5,$6) ON CONFLICT (user_id, preference_key) DO NOTHING
	// Args: [1 theme dark 1 language en]
}

// ExamplePostgres_Builder_subqueries demonstrates building queries with subqueries
func ExamplePostgres_Builder_subqueries() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Subquery in WHERE clause
	subquery, subArgs, err := pg.Builder.
		Select("user_id").
		From("orders").
		Where("total > ?", 1000).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}

	query, args, err := pg.Builder.
		Select("id", "name", "email").
		From("users").
		Where("id IN ("+subquery+")", subArgs...).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// Subquery in SELECT (scalar subquery) - simplified example
	query2, args2, err := pg.Builder.
		Select("p.id", "p.name", "(SELECT AVG(rating) FROM reviews WHERE product_id = p.id) as avg_rating").
		From("products p").
		Where("p.active = ?", true).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: SELECT id, name, email FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > $1)
	// Args: [1000]
	// Query: SELECT p.id, p.name, (SELECT AVG(rating) FROM reviews WHERE product_id = p.id) as avg_rating FROM products p WHERE p.active = $1
	// Args: [true]
}

// ExamplePostgres_Builder_cte demonstrates building queries with Common Table Expressions
func ExamplePostgres_Builder_cte() {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Simple CTE
	query, args, err := pg.Builder.
		Select("*").
		From("monthly_sales").
		PrefixExpr(squirrel.Expr("WITH monthly_sales AS (SELECT DATE_TRUNC('month', order_date) as month, SUM(total) as sales FROM orders GROUP BY DATE_TRUNC('month', order_date))")).
		Where("sales > ?", 10000).
		OrderBy("month").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query, args)

	// Recursive CTE for hierarchical data
	query2, args2, err := pg.Builder.
		Select("*").
		From("category_hierarchy").
		PrefixExpr(squirrel.Expr("WITH RECURSIVE category_hierarchy AS (SELECT id, name, parent_id, 1 as level FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.name, c.parent_id, ch.level + 1 FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id)")).
		Where("level <= ?", 3).
		OrderBy("level", "name").
		ToSql()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\nArgs: %v\n", query2, args2)

	// Output:
	// Query: WITH monthly_sales AS (SELECT DATE_TRUNC('month', order_date) as month, SUM(total) as sales FROM orders GROUP BY DATE_TRUNC('month', order_date)) SELECT * FROM monthly_sales WHERE sales > $1 ORDER BY month
	// Args: [10000]
	// Query: WITH RECURSIVE category_hierarchy AS (SELECT id, name, parent_id, 1 as level FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.name, c.parent_id, ch.level + 1 FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id) SELECT * FROM category_hierarchy WHERE level <= $1 ORDER BY level, name
	// Args: [3]
}
