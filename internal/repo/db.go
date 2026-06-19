package repo

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection to Postgres using the given DSN.
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("repo: connect to postgres: %w", err)
	}
	return db, nil
}

// Migrate applies any *.sql files embedded in fsys that have not yet been
// recorded in the schema_migrations table, in lexical filename order.
func Migrate(ctx context.Context, db *gorm.DB, fsys embed.FS) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("repo: get sql.DB: %w", err)
	}

	if _, err := sqlDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("repo: create schema_migrations table: %w", err)
	}

	entries, err := fsys.ReadDir(".")
	if err != nil {
		return fmt.Errorf("repo: read migrations dir: %w", err)
	}

	var filenames []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			filenames = append(filenames, e.Name())
		}
	}
	sort.Strings(filenames)

	for _, name := range filenames {
		var count int64
		if err := sqlDB.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE filename = $1`, name).Scan(&count); err != nil {
			return fmt.Errorf("repo: check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		contents, err := fsys.ReadFile(name)
		if err != nil {
			return fmt.Errorf("repo: read migration %s: %w", name, err)
		}

		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("repo: begin tx for %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("repo: apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			tx.Rollback()
			return fmt.Errorf("repo: record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("repo: commit migration %s: %w", name, err)
		}
	}

	return nil
}
