package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type FormStatus struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Color        *string `json:"color"`
	DisplayOrder int32   `json:"display_order"`
	IsDefault    bool    `json:"is_default"`
}

func (s *Service) ListFormStatuses(
	ctx context.Context,
	formID string,
	actor uuid.UUID,
) ([]FormStatus, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return nil, err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return nil, ErrValidation
	}
	rows, err := s.Q.ListFormStatuses(ctx, dbUUID(fid))
	if err != nil {
		return nil, err
	}
	out := make([]FormStatus, 0, len(rows))
	for _, row := range rows {
		out = append(out, formStatusFromRow(row))
	}
	return out, nil
}

func (s *Service) CreateFormStatus(
	ctx context.Context,
	formID, name string,
	color *string,
	displayOrder int32,
	isDefault bool,
	actor uuid.UUID,
) (FormStatus, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return FormStatus{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" || displayOrder <= 0 {
		return FormStatus{}, ErrValidation
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return FormStatus{}, ErrValidation
	}

	colorParam := pgtype.Text{Valid: false}
	if color != nil && strings.TrimSpace(*color) != "" {
		colorParam = pgtype.Text{String: strings.TrimSpace(*color), Valid: true}
	}

	row, err := s.Q.CreateFormStatus(ctx, db.CreateFormStatusParams{
		ID:           dbUUID(uuid.New()),
		FormID:       dbUUID(fid),
		Name:         name,
		Color:        colorParam,
		DisplayOrder: displayOrder,
		IsDefault:    isDefault,
	})
	if err != nil {
		return FormStatus{}, err
	}

	if isDefault {
		if _, err := s.Q.SetDefaultFormStatus(ctx, db.SetDefaultFormStatusParams{
			FormID: dbUUID(fid),
			ID:     row.ID,
		}); err != nil {
			return FormStatus{}, err
		}
	}

	return formStatusFromRow(row), nil
}

func (s *Service) UpdateFormStatus(
	ctx context.Context,
	formID, statusID string,
	name, color *string,
	displayOrder *int32,
	actor uuid.UUID,
) (FormStatus, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return FormStatus{}, err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return FormStatus{}, ErrValidation
	}
	sid, err := uuid.Parse(statusID)
	if err != nil {
		return FormStatus{}, ErrValidation
	}

	current, err := s.Q.GetFormStatusByID(ctx, dbUUID(sid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FormStatus{}, ErrForbidden
		}
		return FormStatus{}, err
	}
	if current.FormID != dbUUID(fid) {
		return FormStatus{}, ErrForbidden
	}

	newName := current.Name
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return FormStatus{}, ErrValidation
		}
		newName = trimmed
	}

	newColor := current.Color
	if color != nil {
		trimmed := strings.TrimSpace(*color)
		if trimmed == "" {
			newColor = pgtype.Text{Valid: false}
		} else {
			newColor = pgtype.Text{String: trimmed, Valid: true}
		}
	}

	newOrder := current.DisplayOrder
	if displayOrder != nil {
		if *displayOrder <= 0 {
			return FormStatus{}, ErrValidation
		}
		newOrder = *displayOrder
	}

	row, err := s.Q.UpdateFormStatus(ctx, db.UpdateFormStatusParams{
		ID:           dbUUID(sid),
		Name:         newName,
		Color:        newColor,
		DisplayOrder: newOrder,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FormStatus{}, ErrForbidden
		}
		return FormStatus{}, err
	}
	return formStatusFromRow(row), nil
}

func (s *Service) SetDefaultFormStatus(
	ctx context.Context,
	formID, statusID string,
	actor uuid.UUID,
) (FormStatus, error) {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return FormStatus{}, err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return FormStatus{}, ErrValidation
	}
	sid, err := uuid.Parse(statusID)
	if err != nil {
		return FormStatus{}, ErrValidation
	}

	if err := s.Q.ClearDefaultFormStatus(ctx, dbUUID(fid)); err != nil {
		return FormStatus{}, err
	}
	row, err := s.Q.SetDefaultFormStatus(ctx, db.SetDefaultFormStatusParams{
		FormID: dbUUID(fid),
		ID:     dbUUID(sid),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FormStatus{}, ErrForbidden
		}
		return FormStatus{}, err
	}
	return formStatusFromRow(row), nil
}

func (s *Service) DeleteFormStatus(
	ctx context.Context,
	formID, statusID string,
	actor uuid.UUID,
) error {
	if err := s.RequireEditor(ctx, formID, actor); err != nil {
		return err
	}
	fid, err := uuid.Parse(formID)
	if err != nil {
		return ErrValidation
	}
	sid, err := uuid.Parse(statusID)
	if err != nil {
		return ErrValidation
	}

	status, err := s.Q.GetFormStatusByID(ctx, dbUUID(sid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	if status.FormID != dbUUID(fid) {
		return ErrForbidden
	}
	if status.IsDefault {
		return ErrConflict
	}

	count, err := s.Q.CountTicketsByStatus(ctx, dbUUID(sid))
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrConflict
	}

	if err := s.Q.DeleteFormStatus(ctx, dbUUID(sid)); err != nil {
		return err
	}
	return nil
}

func formStatusFromRow(row db.FormStatus) FormStatus {
	return FormStatus{
		ID:           uuid.UUID(row.ID.Bytes).String(),
		Name:         row.Name,
		Color:        textPtr(row.Color),
		DisplayOrder: row.DisplayOrder,
		IsDefault:    row.IsDefault,
	}
}
