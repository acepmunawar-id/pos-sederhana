package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the pgx5:// driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/migrations"
)

// Connect opens a connection pool and verifies it with a ping.
func Connect(ctx context.Context, url string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = maxConns
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func newMigrator(databaseURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.Schema, ".")
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance("iofs", src, toMigrateURL(databaseURL))
}

// MigrateUp applies all pending migrations.
func MigrateUp(databaseURL string) error {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// MigrateDown rolls back the given number of migrations (all when steps <= 0).
func MigrateDown(databaseURL string, steps int) error {
	m, err := newMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if steps > 0 {
		err = m.Steps(-steps)
	} else {
		err = m.Down()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Seed inserts demo data. The script is idempotent.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, migrations.Seed)
	return err
}

// toMigrateURL switches postgres:// URLs to golang-migrate's pgx5:// scheme.
func toMigrateURL(u string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if len(u) >= len(prefix) && u[:len(prefix)] == prefix {
			return "pgx5://" + u[len(prefix):]
		}
	}
	return u
}
