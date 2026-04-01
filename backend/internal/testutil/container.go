package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres(ctx context.Context) (*pgxpool.Pool, func()) {
	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("testcontainers: start postgres: %v", err))
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("testcontainers: connection string: %v", err))
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("pgxpool: %v", err))
	}

	runMigrations(connStr)

	cleanup := func() {
		pool.Close()
		_ = ctr.Terminate(ctx)
	}
	return pool, cleanup
}

func runMigrations(connStr string) {
	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		panic(fmt.Sprintf("goose open: %v", err))
	}
	defer db.Close()

	migrationsDir := filepath.Join(projectRoot(), "backend", "migrations")
	if err := goose.Up(db, migrationsDir); err != nil {
		panic(fmt.Sprintf("goose up: %v", err))
	}
}

func projectRoot() string {
	_, f, _, _ := runtime.Caller(0)
	// f = .../backend/internal/testutil/container.go
	return filepath.Join(filepath.Dir(f), "..", "..", "..")
}
