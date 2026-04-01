package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.InviteRepository = (*InviteRepository)(nil)

type InviteRepository struct {
	q *db.Queries
}

func NewInviteRepository(pool *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{q: db.New(pool)}
}

func (r *InviteRepository) Create(
	ctx context.Context,
	invite entity.Invite,
) (entity.Invite, error) {
	row, err := r.q.CreateFormInvite(ctx, db.CreateFormInviteParams{
		ID:        toUUID(invite.ID),
		FormID:    toUUID(invite.FormID),
		Email:     invite.Email,
		Role:      db.FormRole(invite.Role),
		InvitedBy: toUUID(invite.InvitedBy),
		ExpiresAt: toTimestamptz(invite.ExpiresAt),
	})
	if err != nil {
		return entity.Invite{}, repoError(err)
	}
	return toInvite(row), nil
}

func (r *InviteRepository) GetForUpdate(ctx context.Context, id uuid.UUID) (entity.Invite, error) {
	row, err := r.q.GetFormInviteForUpdate(ctx, toUUID(id))
	if err != nil {
		return entity.Invite{}, repoError(err)
	}
	return toInvite(row), nil
}

func (r *InviteRepository) Accept(ctx context.Context, id uuid.UUID) (entity.Invite, error) {
	row, err := r.q.AcceptFormInvite(ctx, toUUID(id))
	if err != nil {
		return entity.Invite{}, repoError(err)
	}
	return toInvite(row), nil
}

func (r *InviteRepository) ListActive(
	ctx context.Context,
	formID uuid.UUID,
) ([]entity.Invite, error) {
	rows, err := r.q.ListActiveFormInvites(ctx, toUUID(formID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Invite, len(rows))
	for i, row := range rows {
		result[i] = toInvite(row)
	}
	return result, nil
}

func (r *InviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.DeleteFormInvite(ctx, toUUID(id))
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func toInvite(row db.FormInvite) entity.Invite {
	return entity.Invite{
		ID:         fromUUID(row.ID),
		FormID:     fromUUID(row.FormID),
		Email:      row.Email,
		Role:       string(row.Role),
		InvitedBy:  fromUUID(row.InvitedBy),
		AcceptedAt: fromTimestamptzPtr(row.AcceptedAt),
		ExpiresAt:  fromTimestamptz(row.ExpiresAt),
		CreatedAt:  fromTimestamptz(row.CreatedAt),
	}
}
