package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
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
	r, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: formID,
	})
	if err != nil || r != "admin" {
		return ErrForbidden
	}
	return nil
}

func (s *Service) RequireEditor(ctx context.Context, formID string, actor uuid.UUID) error {
	r, err := s.Q.GetUserFormRole(ctx, db.GetUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: actor, Valid: true},
		FormID: formID,
	})
	if err != nil || (r != "admin" && r != "editor") {
		return ErrForbidden
	}
	return nil
}

func (s *Service) RequireFormAccessForTicket(ctx context.Context, ticketID string, actor uuid.UUID) error {
	uid, err := uuid.Parse(ticketID)
	if err != nil {
		return ErrValidation
	}

	ticket, err := s.Q.GetTicket(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return ErrForbidden
	}

	return s.RequireEditor(ctx, ticket.FormID, actor)
}

func (s *Service) AddMember(ctx context.Context, formID, email, role string) error {
	if role != "admin" && role != "editor" {
		return ErrValidation
	}
	u, err := s.Q.GetUserByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}
	return s.Q.UpsertUserFormRole(ctx, db.UpsertUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: u.ID.Bytes, Valid: true},
		FormID: formID,
		Role:   role,
	})
}

func (s *Service) ChangeRole(ctx context.Context, formID, userID, role string) error {
	if role != "admin" && role != "editor" {
		return ErrValidation
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}
	return s.Q.UpsertUserFormRole(ctx, db.UpsertUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: uid, Valid: true},
		FormID: formID,
		Role:   role,
	})
}

func (s *Service) RemoveMember(ctx context.Context, formID, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrValidation
	}
	return s.Q.DeleteUserFormRole(ctx, db.DeleteUserFormRoleParams{
		UserID: pgtype.UUID{Bytes: uid, Valid: true},
		FormID: formID,
	})
}

func (s *Service) ListMembers(ctx context.Context, formID string) ([]Member, error) {
	rows, err := s.Q.ListFormMembers(ctx, formID)
	if err != nil {
		return nil, err
	}
	out := make([]Member, 0, len(rows))
	for _, r := range rows {
		out = append(out, Member{
			ID:          r.ID.Bytes,
			Email:       r.Email,
			DisplayName: r.DisplayName,
			Role:        r.Role,
		})
	}
	return out, nil
}
