package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RolesService interface {
	AddMember(ctx context.Context, formID, email, role string) error
	ChangeRole(ctx context.Context, formID, userID, role string) error
	RemoveMember(ctx context.Context, formID, userID string) error
	ListMembers(ctx context.Context, formID string) ([]Member, error)
	RequireAdmin(ctx context.Context, formID string, actor uuid.UUID) error
}

type Member struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
}

func (s *Service) RequireAdmin(ctx context.Context, formID string, actor uuid.UUID) error {
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	r, err := s.Roles.GetFormMemberRole(ctx, db.GetFormMemberRoleParams{
		UserID: dbUUID(actor),
		FormID: dbUUID(fid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	if r != db.FormRoleAdmin {
		return ErrForbidden
	}
	return nil
}

func (s *Service) RequireEditor(ctx context.Context, formID string, actor uuid.UUID) error {
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	r, err := s.Roles.GetFormMemberRole(ctx, db.GetFormMemberRoleParams{
		UserID: dbUUID(actor),
		FormID: dbUUID(fid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	if r != db.FormRoleAdmin && r != db.FormRoleEditor {
		return ErrForbidden
	}
	return nil
}

func (s *Service) RequireFormAccessForTicket(
	ctx context.Context,
	ticketID string,
	actor uuid.UUID,
) error {
	uid, err := uuid.Parse(ticketID)
	if err != nil {
		return ErrValidation
	}

	ticket, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}

	return s.RequireEditor(ctx, uuid.UUID(ticket.FormID.Bytes).String(), actor)
}

func (s *Service) AddMember(ctx context.Context, formID, email, role string) error {
	if role != "admin" && role != "editor" {
		return ErrValidation
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	u, err := s.Q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return s.Roles.UpsertFormMember(ctx, db.UpsertFormMemberParams{
		UserID: dbUUID(uuid.UUID(u.ID.Bytes)),
		FormID: dbUUID(fid),
		Role:   db.FormRole(role),
	})
}

func (s *Service) ChangeRole(ctx context.Context, formID, userID, role string) error {
	if role != "admin" && role != "editor" {
		return ErrValidation
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}
	if role != "admin" {
		if err := s.ensureFormKeepsAdmin(ctx, formID, uid); err != nil {
			return err
		}
	}
	return s.Roles.UpsertFormMember(ctx, db.UpsertFormMemberParams{
		UserID: dbUUID(uid),
		FormID: dbUUID(fid),
		Role:   db.FormRole(role),
	})
}

func (s *Service) RemoveMember(ctx context.Context, formID, userID string) error {
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}
	if err := s.ensureFormKeepsAdmin(ctx, formID, uid); err != nil {
		return err
	}
	return s.Roles.DeleteFormMember(ctx, db.DeleteFormMemberParams{
		UserID: dbUUID(uid),
		FormID: dbUUID(fid),
	})
}

func (s *Service) ensureFormKeepsAdmin(ctx context.Context, formID string, userID uuid.UUID) error {
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	role, err := s.Roles.GetFormMemberRole(ctx, db.GetFormMemberRoleParams{
		UserID: dbUUID(userID),
		FormID: dbUUID(fid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if role != db.FormRoleAdmin {
		return nil
	}
	count, err := s.Roles.CountFormAdmins(ctx, dbUUID(fid))
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrValidation
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, formID string) ([]Member, error) {
	fid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}
	rows, err := s.Roles.ListFormMembers(ctx, dbUUID(fid))
	if err != nil {
		return nil, err
	}
	out := make([]Member, 0, len(rows))
	for _, r := range rows {
		out = append(out, Member{
			ID:          r.ID.Bytes,
			Email:       r.Email,
			DisplayName: r.DisplayName,
			Role:        string(r.Role),
		})
	}
	return out, nil
}
