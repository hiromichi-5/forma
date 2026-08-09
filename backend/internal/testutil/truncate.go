package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TruncateAll(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		TRUNCATE
			ticket_notifications,
			ticket_histories,
			tickets,
			form_notification_settings,
			form_statuses,
			form_questions,
			form_invites,
			form_members,
			forms,
			password_reset_tokens,
			email_verification_tokens,
			sessions,
			users
		CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
