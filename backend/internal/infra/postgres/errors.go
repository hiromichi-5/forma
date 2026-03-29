package postgres

import (
	"errors"

	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

func repoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return repository.ErrConflict
	}
	return err
}

func rowsError(n int64) error {
	if n == 0 {
		return repository.ErrNotFound
	}
	return nil
}
