package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RandomUUID() uuid.UUID {
	return uuid.New()
}

func CreateVerifiedUser(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	email, password, displayName string,
) uuid.UUID {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	userID := uuid.New()
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, verified_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, string(hashed), displayName, now, now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return userID
}

func GetEmailVerificationToken(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
) string {
	t.Helper()

	var token string
	err := pool.QueryRow(ctx, `
		SELECT token FROM email_verification_tokens
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&token)
	if err != nil {
		t.Fatalf("get email verification token: %v", err)
	}
	return token
}

func GetPasswordResetToken(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
) string {
	t.Helper()

	var token string
	err := pool.QueryRow(ctx, `
		SELECT token FROM password_reset_tokens
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&token)
	if err != nil {
		t.Fatalf("get password reset token: %v", err)
	}
	return token
}

func CreateForm(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	googleFormID, title string,
	adminUserID uuid.UUID,
) (formID uuid.UUID, defaultStatusID uuid.UUID) {
	t.Helper()

	formID = uuid.New()
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO forms (id, form_id, title, created_at)
		VALUES ($1, $2, $3, $4)
	`, formID, googleFormID, title, now)
	if err != nil {
		t.Fatalf("insert form: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO form_members (user_id, form_id, role, created_at)
		VALUES ($1, $2, 'admin', $3)
	`, adminUserID, formID, now)
	if err != nil {
		t.Fatalf("insert form_member: %v", err)
	}

	defaultStatusID = uuid.New()
	for i, st := range []struct {
		name      string
		color     string
		order     int32
		isDefault bool
	}{
		{"未対応", "#E53935", 1, true},
		{"対応中", "#FB8C00", 2, false},
		{"対応完了", "#43A047", 3, false},
	} {
		sid := defaultStatusID
		if i > 0 {
			sid = uuid.New()
		}
		_, err := pool.Exec(ctx, `
			INSERT INTO form_statuses (id, form_id, name, color, display_order, is_default, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, sid, formID, st.name, st.color, st.order, st.isDefault, now)
		if err != nil {
			t.Fatalf("insert form_status: %v", err)
		}
	}

	return formID, defaultStatusID
}

func AddMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID, formID uuid.UUID,
	role entity.Role,
) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO form_members (user_id, form_id, role, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, formID, role, time.Now())
	if err != nil {
		t.Fatalf("insert form_member: %v", err)
	}
}

func CreateInvite(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	formID, invitedBy uuid.UUID,
	email string,
	role entity.Role,
) uuid.UUID {
	t.Helper()

	inviteID := uuid.New()
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO form_invites (id, form_id, email, role, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, inviteID, formID, email, role, invitedBy, now.Add(24*time.Hour), now)
	if err != nil {
		t.Fatalf("insert form_invite: %v", err)
	}
	return inviteID
}

func CreateTicket(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	formID, statusID uuid.UUID,
	responseID string,
) uuid.UUID {
	t.Helper()
	return CreateTicketWithRespondent(t, ctx, pool, formID, statusID, responseID, "")
}

func CreateTicketWithRespondent(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	formID, statusID uuid.UUID,
	responseID, respondentEmail string,
) uuid.UUID {
	t.Helper()

	var email *string
	if respondentEmail != "" {
		email = &respondentEmail
	}

	ticketID := uuid.New()
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO tickets (id, form_id, response_id, respondent_email, answers, status_id, priority, submitted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'medium', $7, $8)
	`, ticketID, formID, responseID, email, []byte(`{}`), statusID, now, now)
	if err != nil {
		t.Fatalf("insert ticket: %v", err)
	}
	return ticketID
}
