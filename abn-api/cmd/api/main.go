package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andypc/abn-news/internal/api"
	"github.com/andypc/abn-news/internal/cache"
	"github.com/andypc/abn-news/internal/store"
	"github.com/andypc/abn-news/internal/translate"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("ABN News API starting")

	dbURL := getEnv("DATABASE_URL", "postgres://localhost:5432/abnnews")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	port := getEnv("PORT", "8080")
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
	cch := cache.New(redisAddr)
	tr := translate.NewClient()
	router := api.NewRouter(st, cch, tr)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info("API server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("API server stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultMigrationsDir() string {
	// Try relative to CWD first (for go run)
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}
	// Try relative to executable
	exe, err := os.Executable()
	if err != nil {
		return "migrations"
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), "migrations")); err == nil {
		return filepath.Join(filepath.Dir(exe), "migrations")
	}
	// Try relative to binary in Docker
	for _, p := range []string{"/migrations", "../migrations"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "migrations"
}
