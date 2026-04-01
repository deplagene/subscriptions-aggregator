package migrator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

// Migrator - используется для выполнения миграции
type Migrator struct {
	db            *sql.DB
	migrationPath string
}

func NewMigrator(db *sql.DB, migrationPath string) *Migrator {
	return &Migrator{
		db:            db,
		migrationPath: migrationPath,
	}
}

func (m *Migrator) Up(ctx context.Context) error {
	const op = "internal.migrator.Migrator.Up"

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("could not set goose dialect %s: %w", op, err)
	}

	if err := goose.UpContext(ctx, m.db, m.migrationPath); err != nil {
		return fmt.Errorf("could not up migrations %s: %w", op, err)
	}

	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	const op = "internal.migrator.Migrator.Down"

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("could not set goose dialect %s: %w", op, err)
	}

	if err := goose.DownContext(ctx, m.db, m.migrationPath); err != nil {
		return fmt.Errorf("could not down migrations %s: %w", op, err)
	}

	return nil
}
