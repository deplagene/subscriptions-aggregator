// @title Subaggregator API
// @version 1.0
// @description API для управления пользовательскими подписками и расчета их суммарной стоимости.
// @BasePath /
// @schemes http
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/deplagene/subaggregator/docs"
	"github.com/deplagene/subaggregator/internal/config"
	"github.com/deplagene/subaggregator/internal/httpapi"
	httpmiddleware "github.com/deplagene/subaggregator/internal/httpapi/middleware"
	"github.com/deplagene/subaggregator/internal/migrator"
	"github.com/deplagene/subaggregator/internal/postgres"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()

	// кфг
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config.Load: %v", err)
		return
	}

	r := chi.NewRouter()

	// логгер
	logger := logging.NewLogger(
		logging.WithLevel(cfg.Logging.Level),
		logging.WithIsJSON(cfg.Logging.IsJSON),
	)

	ctx = logging.ContextWithLogger(ctx, logger)
	logger.Info("config loaded", "config", cfg.String())

	// middleware
	r.Use(httpmiddleware.LoggerContext(logger))
	r.Use(middleware.Timeout(cfg.API.ReadTimeout))
	r.Use(middleware.Recoverer)
	r.Use(httpmiddleware.RequestLogger())

	// DB
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
	if err != nil {
		logger.Error("main.pgxpool.ParseConfig", "error", err)
		return
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("main.pgxpool.NewWithConfig", "error", err)
		return
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("main.pool.Ping", "error", err)
		return
	}

	migrator := migrator.NewMigrator(stdlib.OpenDB(*pool.Config().ConnConfig), cfg.DB.MigrationPath)
	if err := migrator.Up(ctx); err != nil {
		logger.Error("main.migrator.Up", "error", err)
		return
	}

	// DI
	repo := postgres.NewRepository(pool)
	service := subscriptions.NewSubscriptionsService(repo)
	handlers := httpapi.NewHandler(service)

	handlers.RegisterRoutes(r)

	// server и graceful shutdown
	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.API.Host, strconv.Itoa(cfg.API.Port)),
		Handler:           r,
		ReadHeaderTimeout: cfg.API.ReadHeaderTimeout,
		ReadTimeout:       cfg.API.ReadTimeout,
		WriteTimeout:      cfg.API.WriteTimeout,
	}

	go func() {
		logger.Info("server started", "addr", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("could not start server", "error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(ctx, cfg.API.ShutdownTimeout)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error("could not shutdown server", "error:", err)
	}

	logger.Info("server shutdown")
}
