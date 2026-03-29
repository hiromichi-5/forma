package postgres

import (
	"context"

	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Do(ctx context.Context, fn func(repos repository.Repos) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	q := db.New(tx)
	repos := repository.Repos{
		User:   &UserRepository{q: q},
		Form:   &FormRepository{q: q},
		Member: &MemberRepository{q: q},
		Status: &StatusRepository{q: q},
		Invite: &InviteRepository{q: q},
		Ticket: &TicketRepository{q: q},
	}
	if err := fn(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
