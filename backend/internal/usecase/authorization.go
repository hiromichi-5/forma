package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

type Authorizer struct {
	memberRepo repository.MemberRepository
}

func NewAuthorizer(memberRepo repository.MemberRepository) *Authorizer {
	return &Authorizer{memberRepo: memberRepo}
}

// RequireEditor は非メンバーとロール不足を区別せず RESOURCE_HIDDEN を返す。
// メンバーであることが確認できていないため、フォームの存在自体を隠す。
func (a *Authorizer) RequireEditor(ctx context.Context, formID, userID uuid.UUID) error {
	role, err := a.resolveRole(ctx, formID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return entity.NewError(entity.CodeResourceHidden)
	}
	return nil
}

// RequireAdmin は非メンバーには RESOURCE_HIDDEN、メンバーだが権限が足りない場合は
// FORBIDDEN を返す。メンバーであることは既に分かっているため権限不足を明示する。
func (a *Authorizer) RequireAdmin(ctx context.Context, formID, userID uuid.UUID) error {
	role, err := a.resolveRole(ctx, formID, userID)
	if err != nil {
		return err
	}
	if !role.CanAdmin() {
		return entity.NewError(entity.CodeForbidden)
	}
	return nil
}

func (a *Authorizer) resolveRole(
	ctx context.Context,
	formID, userID uuid.UUID,
) (entity.Role, error) {
	role, err := a.memberRepo.GetRole(ctx, formID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", entity.NewError(entity.CodeResourceHidden)
		}
		return "", err
	}
	return role, nil
}
