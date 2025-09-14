package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListResponses(ctx context.Context, formID string, since *time.Time, actor uuid.UUID) ([]db.Response, error) {
	// If formID is provided, check access to that specific form
	if formID != "" {
		if err := s.RequireEditor(ctx, formID, actor); err != nil {
			return nil, err
		}
		var sinceTs pgtype.Timestamptz
		if since != nil {
			sinceTs = pgtype.Timestamptz{Time: *since, Valid: true}
		}
		return s.Q.ListResponses(ctx, db.ListResponsesParams{
			Column1: formID,
			Column2: sinceTs,
		})
	}

	// If no formID specified, return empty list for security
	// In a multi-tenant system, this would filter by user's accessible forms
	return []db.Response{}, nil
}
