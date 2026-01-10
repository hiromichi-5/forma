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

func (s *Service) CreateInvite(ctx context.Context, formID, email, role string, actor uuid.UUID) (db.FormInvite, error) {
	if role != "admin" && role != "editor" {
		return db.FormInvite{}, ErrValidation
	}
	if err := s.RequireAdmin(ctx, formID, actor); err != nil {
		return db.FormInvite{}, err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return db.FormInvite{}, ErrValidation
	}

	expiresAt := pgtype.Timestamptz{Time: s.nowTime().Add(inviteTTL), Valid: true}

	invite, err := s.Invites.CreateFormInvite(ctx, db.CreateFormInviteParams{
		ID:        dbUUID(uuid.New()),
		FormID:    dbUUID(fid),
		Email:     email,
		Role:      role,
		InvitedBy: dbUUID(actor),
		ExpiresAt: expiresAt,
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
	fid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}

	return s.Invites.ListActiveFormInvites(ctx, dbUUID(fid))
}

func (s *Service) DeleteInvite(ctx context.Context, formID, inviteID string, actor uuid.UUID) error {
	if err := s.RequireAdmin(ctx, formID, actor); err != nil {
		return err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	iid, err := uuid.Parse(inviteID)
	if err != nil {
		return ErrValidation
	}
	inv, err := s.Invites.GetFormInviteForUpdate(ctx, dbUUID(iid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}
	if inv.FormID != dbUUID(fid) {
		return ErrInviteNotFound
	}
	return s.Invites.DeleteFormInvite(ctx, dbUUID(iid))
}

func (s *Service) AcceptInvite(ctx context.Context, inviteID string, actor uuid.UUID) error {
	invID, err := uuid.Parse(inviteID)
	if err != nil {
		return ErrValidation
	}

	invite, err := s.Invites.GetFormInviteForUpdate(ctx, dbUUID(invID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}

	if invite.AcceptedAt.Valid {
		return ErrInviteNotFound
	}
	if !invite.ExpiresAt.Valid || !invite.ExpiresAt.Time.After(s.nowTime()) {
		return ErrInviteExpired
	}

	user, err := s.Users.GetUserByID(ctx, dbUUID(actor))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	if user.Email != invite.Email {
		return ErrForbidden
	}

	_, err = s.Invites.AcceptFormInvite(ctx, dbUUID(invID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}

	return s.Roles.UpsertFormMember(ctx, db.UpsertFormMemberParams{
		UserID: dbUUID(actor),
		FormID: invite.FormID,
		Role:   string(invite.Role),
	})
}
