package postgres_test

import (
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/rdashevsky/go-pkgs/postgres"
)

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

// BenchmarkSquirrelSelect benchmarks different SELECT query complexities
func BenchmarkSquirrelSelect(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.Run("SimpleSelect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select("id, name").
				From("users").
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SelectWithWhere", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select("id, name, email").
				From("users").
				Where(squirrel.Eq{"active": true}).
				Where("age > ?", 18).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexSelectWithJoins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select("u.id", "u.name", "p.title", "c.content").
				From("users u").
				Join("posts p ON u.id = p.user_id").
				LeftJoin("comments c ON p.id = c.post_id").
				Where(squirrel.And{
					squirrel.Eq{"u.active": true},
					squirrel.Gt{"p.created_at": time.Now().Add(-30 * 24 * time.Hour)},
				}).
				GroupBy("u.id", "u.name", "p.title", "c.content").
				Having("COUNT(c.id) > ?", 0).
				OrderBy("u.name", "p.created_at DESC").
				Limit(100).
				Offset(50).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSquirrelInsert benchmarks different INSERT operations
func BenchmarkSquirrelInsert(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.Run("SingleInsert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Insert("users").
				Columns("name", "email", "active").
				Values("John Doe", "john@example.com", true).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("BatchInsert10Rows", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			insertBuilder := pg.Builder.Insert("users").Columns("name", "email", "active")
			for j := 0; j < 10; j++ {
				insertBuilder = insertBuilder.Values("User", "user@example.com", true)
			}
			_, _, err := insertBuilder.ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("BatchInsert100Rows", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			insertBuilder := pg.Builder.Insert("users").Columns("name", "email", "active")
			for j := 0; j < 100; j++ {
				insertBuilder = insertBuilder.Values("User", "user@example.com", true)
			}
			_, _, err := insertBuilder.ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpsertWithConflict", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Insert("users").
				Columns("email", "name", "updated_at").
				Values("john@example.com", "John Doe", "NOW()").
				Suffix("ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at").
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSquirrelUpdate benchmarks UPDATE operations
func BenchmarkSquirrelUpdate(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.Run("SimpleUpdate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Update("users").
				Set("name", "Updated Name").
				Where(squirrel.Eq{"id": 123}).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpdateMultipleFields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Update("users").
				Set("name", "Updated Name").
				Set("email", "updated@example.com").
				Set("active", false).
				Set("updated_at", "NOW()").
				Where(squirrel.Eq{"id": 123}).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpdateWithComplexWhere", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Update("users").
				Set("last_login", "NOW()").
				Where(squirrel.And{
					squirrel.Eq{"active": true},
					squirrel.Lt{"last_login": time.Now().Add(-7 * 24 * time.Hour)},
					squirrel.Eq{"role": []interface{}{"user", "admin"}},
				}).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSquirrelDelete benchmarks DELETE operations
func BenchmarkSquirrelDelete(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.Run("SimpleDelete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Delete("users").
				Where(squirrel.Eq{"id": 123}).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("DeleteWithMultipleConditions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Delete("users").
				Where("active = ?", false).
				Where("created_at < ?", time.Now().Add(-30*24*time.Hour)).
				Where("login_count = ?", 0).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkNewWithOptions benchmarks connection creation with different option combinations
func BenchmarkNewWithOptions(b *testing.B) {
	b.Run("NoOptions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pg, _ := postgres.New(
				"postgres://user:pass@127.0.0.1:65432/db",
				postgres.ConnAttempts(1),
			)
			if pg != nil {
				pg.Close()
			}
		}
	})

	b.Run("SingleOption", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pg, _ := postgres.New(
				"postgres://user:pass@127.0.0.1:65432/db",
				postgres.MaxPoolSize(10),
				postgres.ConnAttempts(1),
			)
			if pg != nil {
				pg.Close()
			}
		}
	})

	b.Run("MultipleOptions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pg, _ := postgres.New(
				"postgres://user:pass@127.0.0.1:65432/db",
				postgres.MaxPoolSize(25),
				postgres.ConnTimeout(5*time.Second),
				postgres.ConnAttempts(1),
			)
			if pg != nil {
				pg.Close()
			}
		}
	})
}

// BenchmarkSquirrelPlaceholderFormats benchmarks different placeholder formats
func BenchmarkSquirrelPlaceholderFormats(b *testing.B) {
	b.Run("DollarPlaceholder", func(b *testing.B) {
		pg := &postgres.Postgres{}
		pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select("id, name").
				From("users").
				Where(squirrel.Eq{"active": true}).
				Where("age > ?", 18).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("QuestionPlaceholder", func(b *testing.B) {
		builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

		for i := 0; i < b.N; i++ {
			_, _, err := builder.
				Select("id, name").
				From("users").
				Where(squirrel.Eq{"active": true}).
				Where("age > ?", 18).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSquirrelComplexQueries benchmarks real-world complex queries
func BenchmarkSquirrelComplexQueries(b *testing.B) {
	pg := &postgres.Postgres{}
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	b.Run("ECommerceOrdersQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select(
					"o.id",
					"o.order_date",
					"u.name as customer_name",
					"u.email",
					"SUM(oi.quantity * p.price) as total_amount",
					"COUNT(oi.id) as item_count",
				).
				From("orders o").
				Join("users u ON o.user_id = u.id").
				Join("order_items oi ON o.id = oi.order_id").
				Join("products p ON oi.product_id = p.id").
				Where(squirrel.And{
					squirrel.Eq{"o.status": "completed"},
					squirrel.GtOrEq{"o.order_date": time.Now().Add(-30 * 24 * time.Hour)},
					squirrel.Eq{"u.active": true},
				}).
				GroupBy("o.id", "o.order_date", "u.name", "u.email").
				Having("SUM(oi.quantity * p.price) > ?", 100.00).
				OrderBy("total_amount DESC", "o.order_date DESC").
				Limit(50).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("BlogAnalyticsQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := pg.Builder.
				Select(
					"p.id",
					"p.title",
					"u.name as author",
					"COUNT(DISTINCT c.id) as comment_count",
					"COUNT(DISTINCT l.id) as like_count",
					"COUNT(DISTINCT v.id) as view_count",
				).
				From("posts p").
				Join("users u ON p.author_id = u.id").
				LeftJoin("comments c ON p.id = c.post_id AND c.status = 'approved'").
				LeftJoin("likes l ON p.id = l.post_id").
				LeftJoin("views v ON p.id = v.post_id AND v.created_at >= ?", time.Now().Add(-7*24*time.Hour)).
				Where(squirrel.And{
					squirrel.Eq{"p.published": true},
					squirrel.GtOrEq{"p.created_at": time.Now().Add(-90 * 24 * time.Hour)},
				}).
				GroupBy("p.id", "p.title", "u.name").
				Having("COUNT(DISTINCT v.id) > ?", 10).
				OrderBy("view_count DESC", "comment_count DESC", "like_count DESC").
				Limit(20).
				ToSql()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
