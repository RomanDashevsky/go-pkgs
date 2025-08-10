// Package postgres provides PostgreSQL database connectivity with connection pooling
// and query building capabilities using pgx driver and Squirrel query builder.
package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

// Postgres represents a PostgreSQL database connection with connection pooling.
// It includes a Squirrel query builder configured for PostgreSQL.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	// Builder is a Squirrel query builder configured with PostgreSQL dollar placeholders.
	Builder squirrel.StatementBuilderType
	// Pool is the underlying pgx connection pool.
	Pool *pgxpool.Pool
}

// New creates a new PostgreSQL connection with the given database URL and options.
// It attempts to establish connection with retry logic and configures Squirrel query builder.
// Default configuration: max pool size 1, 10 connection attempts, 1 second timeout between attempts.
//
// Example:
//
//	pg, err := postgres.New("postgres://user:pass@localhost/db",
//	    postgres.MaxPoolSize(10),
//	    postgres.ConnAttempts(5),
//	)
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize) // #nosec G115 -- maxPoolSize is controlled and validated

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close gracefully closes the database connection pool.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
