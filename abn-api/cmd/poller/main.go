package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andypc/abn-news/internal/rss"
	"github.com/andypc/abn-news/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("ABN News RSS Poller starting")

	dbURL := getEnv("DATABASE_URL", "postgres://localhost:5432/abnnews")
	migrationsDir := getEnv("MIGRATIONS_DIR", defaultMigrationsDir())

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := store.RunMigrations(context.Background(), pool, migrationsDir); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	st := store.New(pool)
	fetcher := rss.NewFetcher(10 * time.Second)
	poller := rss.NewPoller(st, fetcher, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	poller.Start(ctx)
	logger.Info("Poller started, waiting for signals...")

	<-ctx.Done()
	logger.Info("Shutting down poller...")
	poller.Stop()
	logger.Info("Poller stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultMigrationsDir() string {
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}
	exe, err := os.Executable()
	if err != nil {
		return "migrations"
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), "migrations")); err == nil {
		return filepath.Join(filepath.Dir(exe), "migrations")
	}
	for _, p := range []string{"/migrations", "../migrations"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "migrations"
}
