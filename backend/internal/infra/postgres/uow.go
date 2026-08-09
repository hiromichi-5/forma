package postgres

import (
	"context"

	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type unitOfWork[T any] struct {
	pool    *pgxpool.Pool
	factory func(q *db.Queries) T
}

func (u *unitOfWork[T]) Do(ctx context.Context, fn func(repos T) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	q := db.New(tx)
	repos := u.factory(q)
	if err := fn(repos); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func NewAuthUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.AuthRepos] {
	return &unitOfWork[repository.AuthRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.AuthRepos {
			return repository.AuthRepos{User: &UserRepository{q: q}}
		},
	}
}

func NewFormUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.FormRepos] {
	return &unitOfWork[repository.FormRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.FormRepos {
			return repository.FormRepos{
				Form:   &FormRepository{q: q},
				Member: &MemberRepository{q: q},
				Status: &StatusRepository{q: q},
			}
		},
	}
}

func NewInviteUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.InviteRepos] {
	return &unitOfWork[repository.InviteRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.InviteRepos {
			return repository.InviteRepos{
				Invite: &InviteRepository{q: q},
				User:   &UserRepository{q: q},
				Member: &MemberRepository{q: q},
			}
		},
	}
}

func NewStatusUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.StatusRepos] {
	return &unitOfWork[repository.StatusRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.StatusRepos {
			return repository.StatusRepos{Status: &StatusRepository{q: q}}
		},
	}
}

func NewNotificationUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.NotificationRepos] {
	return &unitOfWork[repository.NotificationRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.NotificationRepos {
			return repository.NotificationRepos{Notification: &NotificationRepository{q: q}}
		},
	}
}

func NewTicketUoW(pool *pgxpool.Pool) repository.UnitOfWork[repository.TicketRepos] {
	return &unitOfWork[repository.TicketRepos]{
		pool: pool,
		factory: func(q *db.Queries) repository.TicketRepos {
			return repository.TicketRepos{
				Ticket: &TicketRepository{q: q},
				Status: &StatusRepository{q: q},
				User:   &UserRepository{q: q},
			}
		},
	}
}
