package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrations: set dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("migrations: up: %w", err)
	}

	return nil
}
