package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"auth-backend/internal/app"
)

const shutdownTimeout = 10 * time.Second

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := app.Config{
		DatabaseURL:     mustEnv("DATABASE_URL"),
		MigrationsPath:  env("MIGRATIONS_PATH", "migrations"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379/0"),
		RedisKeyPrefix:  env("REDIS_KEY_PREFIX", "auth:"),
		JWTSecret:       mustEnv("JWT_SECRET"),
		AccessTTL:       envDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTTL:      envDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		PermCacheTTL:    envDuration("PERMISSIONS_CACHE_TTL", 5*time.Minute),
		LoginRateLimit:  int64(envInt("LOGIN_RATE_LIMIT", 5)),
		LoginRateWindow: envDuration("LOGIN_RATE_WINDOW", 15*time.Minute),
		AdminEmail:      env("ADMIN_EMAIL", "admin@local.dev"),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
		CookieSecure:    env("COOKIE_SECURE", "false") == "true",
	}
	port := env("PORT", "8080")

	application, err := app.New(cfg)
	if err != nil {
		fatal("startup failed", err)
	}
	defer application.Close()
	slog.Info("migrations applied")

	if err := application.SeedAdmin(context.Background()); err != nil {
		fatal("admin seed failed", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- application.Fiber.Listen(":" + port)
	}()
	slog.Info("server listening", "port", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		fatal("server stopped unexpectedly", err)
	case sig := <-quit:
		slog.Info("shutting down", "signal", sig.String())
		if err := application.Fiber.ShutdownWithTimeout(shutdownTimeout); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
		}
	}
}

func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		slog.Warn("invalid duration env var, using default", "key", key, "default", fallback.String())
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		slog.Warn("invalid int env var, using default", "key", key, "default", fallback)
	}
	return fallback
}
