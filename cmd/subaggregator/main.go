package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deplagene/subaggregator/internal/httpapi"
	"github.com/deplagene/subaggregator/internal/migrator"
	"github.com/deplagene/subaggregator/internal/postgres"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()
	r := chi.NewRouter()

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
	repo := postgres.NewRepository(pool)
	service := subscriptions.NewSubscriptionsService(repo)
	handlers := httpapi.NewHandler(service)

	handlers.RegisterRoutes(r)

	// API

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful Shutdown

	go func() {
		logger.Info("server started at port", "port", "todo")
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("could not start server", "error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error("could not shutdown server", "error:", err)
	}

	logger.Info("server shutdown")
}
