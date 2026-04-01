package main

import (
	"context"

	"github.com/deplagene/subaggregator/internal/migrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()

	// кфг файл

	// логгер
	logger := logging.NewLogger(
		logging.WithLevel("info"),
		logging.WithIsJSON(true),
	)

	ctx = logging.ContextWithLogger(ctx, logger)

	// БД
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/subaggregator")
	if err != nil {
		logger.Error("main.pgxpool.New", "error", err)
		return
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("main.pool.Ping", "error", err)
		return
	}

	migrator := migrator.NewMigrator(stdlib.OpenDB(*pool.Config().ConnConfig), "")
	if err := migrator.Up(ctx); err != nil {
		logger.Error("main.migrator.Up", "error", err)
		return
	}

	// DI

	// API

	// Graceful Shutdown
}
