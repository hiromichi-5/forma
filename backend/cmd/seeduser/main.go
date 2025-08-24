package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"

	"github.com/hiromichi-5/forma/backend/internal/db"
)

func main() {
	viper.AutomaticEnv()

	pgDSN := viper.GetString("PG_DSN")
	if pgDSN == "" {
		log.Fatal("PG_DSN required")
	}

	email := os.Getenv("SEED_EMAIL")
	pass := os.Getenv("SEED_PASSWORD")
	if email == "" || pass == "" {
		log.Fatal("SEED_EMAIL and SEED_PASSWORD required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()
	q := db.New(pool)

	u, err := q.GetUserByEmail(ctx, email)
	if err == nil {
		fmt.Printf("user already exists: %s (id=%s)\n", u.Email, u.ID.Bytes)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	uid := uuid.New()
	newU, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:           pgtype.UUID{Bytes: uid, Valid: true},
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Fatalf("insert: %v", err)
	}
	fmt.Printf("created user: %s (id=%s)\n", newU.Email, newU.ID.Bytes)
}
