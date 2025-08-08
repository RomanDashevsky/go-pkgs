// Package goose provides database migration utilities built on top of pressly/goose.
// It offers functions to check migration status and validate database schema versions.
package goose

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rdashevsky/go-pkgs/logger"
)

// CheckMigrationStatus checks if the current database migration version matches the expected version.
// Returns the current version and an error if versions don't match or if there's a database error.
func CheckMigrationStatus(pool *pgxpool.Pool, expectedVersion int64, l logger.LoggerI) (int64, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			l.Error("CheckMigrationStatus - DB close failed: %v", err)
		}
	}(db)

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get database version: %w", err)
	}

	if currentVersion != expectedVersion {
		return currentVersion, fmt.Errorf("database schema version %d does not match expected %d", currentVersion, expectedVersion)
	}

	l.Info("Migrations are up to date: %d", currentVersion)
	return currentVersion, nil
}
