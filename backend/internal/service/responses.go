package service

import (
	"context"
	"time"

	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListResponses(ctx context.Context, formID string, since *time.Time) ([]db.Response, error) {
	var sinceTs pgtype.Timestamptz
	if since != nil {
		sinceTs = pgtype.Timestamptz{Time: *since, Valid: true}
	}
	return s.Q.ListResponses(ctx, db.ListResponsesParams{
		Column1: formID,
		Column2: sinceTs,
	})
}
