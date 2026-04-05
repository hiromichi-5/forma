package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	pgDSN := os.Getenv("PG_DSN")
	if pgDSN == "" {
		return fmt.Errorf("PG_DSN required")
	}

	migrationDir := os.Getenv("MIGRATION_DIR")
	if migrationDir == "" {
		migrationDir = "/migrations"
	}

	ctx := context.Background()

	db, err := sql.Open("pgx", pgDSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close() //nolint:errcheck

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	log.Printf( //nolint:gosec
		"running migrations from %s",
		migrationDir,
	)
	if err := goose.UpContext(ctx, db, migrationDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Println("migrations completed successfully")
	return nil
}
