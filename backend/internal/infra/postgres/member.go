package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	db "github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/hiromichi-5/forma/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.MemberRepository = (*MemberRepository)(nil)

type MemberRepository struct {
	q *db.Queries
}

func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{q: db.New(pool)}
}

func (r *MemberRepository) Upsert(
	ctx context.Context,
	userID uuid.UUID,
	formID uuid.UUID,
	role entity.Role,
) error {
	return r.q.UpsertFormMember(ctx, db.UpsertFormMemberParams{
		UserID: toUUID(userID),
		FormID: toUUID(formID),
		Role:   db.FormRole(role),
	})
}

func (r *MemberRepository) Delete(ctx context.Context, userID uuid.UUID, formID uuid.UUID) error {
	n, err := r.q.DeleteFormMember(ctx, db.DeleteFormMemberParams{
		UserID: toUUID(userID),
		FormID: toUUID(formID),
	})
	if err != nil {
		return repoError(err)
	}
	return rowsError(n)
}

func (r *MemberRepository) GetRole(
	ctx context.Context,
	userID uuid.UUID,
	formID uuid.UUID,
) (entity.Role, error) {
	role, err := r.q.GetFormMemberRole(ctx, db.GetFormMemberRoleParams{
		UserID: toUUID(userID),
		FormID: toUUID(formID),
	})
	if err != nil {
		return "", repoError(err)
	}
	return entity.Role(role), nil
}

func (r *MemberRepository) List(ctx context.Context, formID uuid.UUID) ([]entity.Member, error) {
	rows, err := r.q.ListFormMembers(ctx, toUUID(formID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Member, len(rows))
	for i, row := range rows {
		result[i] = entity.Member{
			UserRef: entity.UserRef{
				ID:          fromUUID(row.ID),
				Email:       row.Email,
				DisplayName: row.DisplayName,
			},
			Role: entity.Role(row.Role),
		}
	}
	return result, nil
}

func (r *MemberRepository) CountAdmins(ctx context.Context, formID uuid.UUID) (int64, error) {
	return r.q.CountFormAdmins(ctx, toUUID(formID))
}

func (r *MemberRepository) ListAccessibleForms(
	ctx context.Context,
	userID uuid.UUID,
) ([]entity.Form, error) {
	rows, err := r.q.ListUserAccessibleForms(ctx, toUUID(userID))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Form, len(rows))
	for i, row := range rows {
		result[i] = entity.Form{
			ID:       fromUUID(row.ID),
			FormID:   row.FormID,
			Title:    row.Title,
			SyncedAt: fromTimestamptzPtr(row.SyncedAt),
		}
	}
	return result, nil
}
