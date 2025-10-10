package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const inviteTTL = 7 * 24 * time.Hour

func (s *Service) CreateInvite(ctx context.Context, formID string, actor uuid.UUID) (db.FormInvite, error) {
	if err := s.RequireAdmin(ctx, formID, actor); err != nil {
		return db.FormInvite{}, err
	}

	code, err := s.newInviteCode()
	if err != nil {
		return db.FormInvite{}, ErrCodeGeneration
	}

	expiresAt := pgtype.Timestamptz{Time: s.nowTime().Add(inviteTTL), Valid: true}

	invite, err := s.Invites.CreateFormInvite(ctx, db.CreateFormInviteParams{
		Code:      code,
		FormID:    formID,
		Role:      "editor",
		ExpiresAt: expiresAt,
		CreatedBy: pgtype.UUID{Bytes: actor, Valid: true},
	})
	if err != nil {
		return db.FormInvite{}, err
	}

	return invite, nil
}

func (s *Service) ListInvites(ctx context.Context, formID string, actor uuid.UUID) ([]db.FormInvite, error) {
	if err := s.RequireAdmin(ctx, formID, actor); err != nil {
		return nil, err
	}

	return s.Invites.ListActiveFormInvites(ctx, db.ListActiveFormInvitesParams{
		FormID:    formID,
		ExpiresAt: pgtype.Timestamptz{Time: s.nowTime(), Valid: true},
	})
}

func (s *Service) RevokeInvite(ctx context.Context, formID, code string, actor uuid.UUID) (db.FormInvite, error) {
	if err := s.RequireAdmin(ctx, formID, actor); err != nil {
		return db.FormInvite{}, err
	}

	invite, err := s.Invites.RevokeFormInvite(ctx, db.RevokeFormInviteParams{
		FormID: formID,
		Code:   code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.FormInvite{}, ErrInviteNotFound
		}
		return db.FormInvite{}, err
	}
	return invite, nil
}

func (s *Service) AcceptInvite(ctx context.Context, code string, actor uuid.UUID) error {
	invite, err := s.Invites.GetFormInviteForUpdate(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}

	if invite.Revoked {
		return ErrInviteRevoked
	}
	if !invite.ExpiresAt.Valid || !invite.ExpiresAt.Time.After(s.nowTime()) {
		return ErrInviteExpired
	}

	role, err := s.Roles.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: invite.FormID,
	})
	if err == nil {
		if role == "admin" || role == "editor" {
			return ErrAlreadyMember
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if err := s.Roles.UpsertUserFormRole(ctx, db.UpsertUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: invite.FormID,
		Role:   invite.Role,
	}); err != nil {
		return err
	}

	_, err = s.Invites.RevokeFormInvite(ctx, db.RevokeFormInviteParams{
		FormID: invite.FormID,
		Code:   invite.Code,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}
