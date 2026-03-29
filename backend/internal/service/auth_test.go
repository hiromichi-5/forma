package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type fakeAuthStore struct {
	users       map[string]db.GetUserByEmailRow
	usersByID   map[uuid.UUID]db.GetUserByEmailRow
	sessions    map[uuid.UUID]db.Session
	emailTokens map[string]db.EmailVerificationToken
	resetTokens map[string]db.PasswordResetToken
	now         time.Time
}

func (f *fakeAuthStore) GetUserByEmail(
	_ context.Context,
	email string,
) (db.GetUserByEmailRow, error) {
	u, ok := f.users[email]
	if !ok {
		return db.GetUserByEmailRow{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeAuthStore) CreateUser(
	_ context.Context,
	arg db.CreateUserParams,
) (db.CreateUserRow, error) {
	if f.users == nil {
		f.users = map[string]db.GetUserByEmailRow{}
	}
	if f.usersByID == nil {
		f.usersByID = map[uuid.UUID]db.GetUserByEmailRow{}
	}
	if _, ok := f.users[arg.Email]; ok {
		return db.CreateUserRow{}, errors.New("duplicate")
	}
	u := db.GetUserByEmailRow{
		ID:           pgUUID(arg.ID.Bytes),
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: f.now, Valid: true},
		DisplayName:  arg.DisplayName,
		VerifiedAt:   arg.VerifiedAt,
	}
	f.users[arg.Email] = u
	f.usersByID[arg.ID.Bytes] = u
	return db.CreateUserRow(u), nil
}

func (f *fakeAuthStore) CreateSession(
	_ context.Context,
	arg db.CreateSessionParams,
) (db.Session, error) {
	if f.sessions == nil {
		f.sessions = map[uuid.UUID]db.Session{}
	}
	s := db.Session{
		ID:        arg.ID,
		UserID:    arg.UserID,
		CreatedAt: pgtype.Timestamptz{Time: f.now, Valid: true},
	}
	f.sessions[arg.ID.Bytes] = s
	return s, nil
}

func (f *fakeAuthStore) DeleteSession(_ context.Context, id pgtype.UUID) (int64, error) {
	if f.sessions == nil {
		return 0, nil
	}
	if _, ok := f.sessions[id.Bytes]; !ok {
		return 0, nil
	}
	delete(f.sessions, id.Bytes)
	return 1, nil
}

func (f *fakeAuthStore) CreateEmailVerificationToken(
	_ context.Context,
	arg db.CreateEmailVerificationTokenParams,
) (db.EmailVerificationToken, error) {
	if f.emailTokens == nil {
		f.emailTokens = map[string]db.EmailVerificationToken{}
	}
	t := db.EmailVerificationToken{
		ID:        arg.ID,
		UserID:    arg.UserID,
		Token:     arg.Token,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: pgtype.Timestamptz{Time: f.now, Valid: true},
	}
	f.emailTokens[arg.Token] = t
	return t, nil
}

func (f *fakeAuthStore) GetEmailVerificationTokenByToken(
	_ context.Context,
	token string,
) (db.EmailVerificationToken, error) {
	t, ok := f.emailTokens[token]
	if !ok {
		return db.EmailVerificationToken{}, pgx.ErrNoRows
	}
	return t, nil
}

func (f *fakeAuthStore) UseEmailVerificationToken(
	_ context.Context,
	id pgtype.UUID,
) (int64, error) {
	for k, t := range f.emailTokens {
		if t.ID == id {
			t.UsedAt = pgtype.Timestamptz{Time: f.now, Valid: true}
			f.emailTokens[k] = t
			return 1, nil
		}
	}
	return 0, nil
}

func (f *fakeAuthStore) DeleteEmailVerificationTokensByUser(
	_ context.Context,
	userID pgtype.UUID,
) error {
	for k, t := range f.emailTokens {
		if t.UserID == userID {
			delete(f.emailTokens, k)
		}
	}
	return nil
}

func (f *fakeAuthStore) SetUserVerifiedAt(_ context.Context, arg db.SetUserVerifiedAtParams) error {
	u, ok := f.usersByID[arg.ID.Bytes]
	if !ok {
		return pgx.ErrNoRows
	}
	u.VerifiedAt = arg.VerifiedAt
	f.usersByID[arg.ID.Bytes] = u
	f.users[u.Email] = u
	return nil
}

func (f *fakeAuthStore) CreatePasswordResetToken(
	_ context.Context,
	arg db.CreatePasswordResetTokenParams,
) (db.PasswordResetToken, error) {
	if f.resetTokens == nil {
		f.resetTokens = map[string]db.PasswordResetToken{}
	}
	t := db.PasswordResetToken{
		ID:        arg.ID,
		UserID:    arg.UserID,
		Token:     arg.Token,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: pgtype.Timestamptz{Time: f.now, Valid: true},
	}
	f.resetTokens[arg.Token] = t
	return t, nil
}

func (f *fakeAuthStore) GetPasswordResetTokenByToken(
	_ context.Context,
	token string,
) (db.PasswordResetToken, error) {
	t, ok := f.resetTokens[token]
	if !ok {
		return db.PasswordResetToken{}, pgx.ErrNoRows
	}
	return t, nil
}

func (f *fakeAuthStore) UsePasswordResetToken(_ context.Context, id pgtype.UUID) (int64, error) {
	for k, t := range f.resetTokens {
		if t.ID == id {
			t.UsedAt = pgtype.Timestamptz{Time: f.now, Valid: true}
			f.resetTokens[k] = t
			return 1, nil
		}
	}
	return 0, nil
}

func (f *fakeAuthStore) DeletePasswordResetTokensByUser(
	_ context.Context,
	userID pgtype.UUID,
) error {
	for k, t := range f.resetTokens {
		if t.UserID == userID {
			delete(f.resetTokens, k)
		}
	}
	return nil
}

func (f *fakeAuthStore) UpdateUserPasswordHash(
	_ context.Context,
	arg db.UpdateUserPasswordHashParams,
) error {
	u, ok := f.usersByID[arg.ID.Bytes]
	if !ok {
		return pgx.ErrNoRows
	}
	u.PasswordHash = arg.PasswordHash
	f.usersByID[arg.ID.Bytes] = u
	f.users[u.Email] = u
	return nil
}

func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func mustHash() string {
	b, err := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestAuthenticate_Success(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	store := &fakeAuthStore{now: time.Unix(1_700_000_000, 0)}
	store.users = map[string]db.GetUserByEmailRow{
		"a@example.com": {
			ID:           pgUUID(uid),
			Email:        "a@example.com",
			PasswordHash: mustHash(),
			CreatedAt:    pgtype.Timestamptz{Time: store.now, Valid: true},
			DisplayName:  "Test User",
			VerifiedAt:   pgtype.Timestamptz{Time: store.now, Valid: true},
		},
	}
	store.usersByID = map[uuid.UUID]db.GetUserByEmailRow{uid: store.users["a@example.com"]}

	s := NewAuthService(store)
	s.now = func() time.Time { return store.now }

	sessionID, err := s.Authenticate(context.Background(), "a@example.com", "pass123")
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if sessionID == "" {
		t.Fatalf("want non-empty sessionID")
	}
}

func TestAuthenticate_EmailNotVerified(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	store := &fakeAuthStore{now: time.Unix(1_700_000_000, 0)}
	store.users = map[string]db.GetUserByEmailRow{
		"a@example.com": {
			ID:           pgUUID(uid),
			Email:        "a@example.com",
			PasswordHash: mustHash(),
			CreatedAt:    pgtype.Timestamptz{Time: store.now, Valid: true},
			DisplayName:  "Test User",
			VerifiedAt:   pgtype.Timestamptz{},
		},
	}
	store.usersByID = map[uuid.UUID]db.GetUserByEmailRow{uid: store.users["a@example.com"]}

	s := NewAuthService(store)
	_, err := s.Authenticate(context.Background(), "a@example.com", "pass123")
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("want ErrEmailNotVerified, got %v", err)
	}
}

func TestSignup_CreatesVerificationToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeAuthStore{now: now}
	s := NewAuthService(store)
	s.now = func() time.Time { return now }
	s.generateToken = func() (string, error) { return "tok", nil }

	uid, err := s.Signup(context.Background(), "new@example.com", "pass12345", "Test User")
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if uid == "" {
		t.Fatalf("want non-empty userID")
	}
	if store.emailTokens["tok"].ExpiresAt.Time != now.Add(24*time.Hour) {
		t.Fatalf("unexpected expires_at")
	}
}

func TestVerifyEmail_SetsVerifiedAt(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	store := &fakeAuthStore{now: now}
	store.users = map[string]db.GetUserByEmailRow{
		"a@example.com": {
			ID:           pgUUID(uid),
			Email:        "a@example.com",
			PasswordHash: mustHash(),
			CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			DisplayName:  "Test User",
			VerifiedAt:   pgtype.Timestamptz{},
		},
	}
	store.usersByID = map[uuid.UUID]db.GetUserByEmailRow{uid: store.users["a@example.com"]}
	store.emailTokens = map[string]db.EmailVerificationToken{
		"tok": {
			ID:     pgUUID(uuid.New()),
			UserID: pgUUID(uid),
		},
	}

	s := NewAuthService(store)
	s.now = func() time.Time { return now }

	if err := s.VerifyEmail(context.Background(), "tok"); err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	u := store.usersByID[uid]
	if !u.VerifiedAt.Valid {
		t.Fatalf("want verified_at to be set")
	}
}

func TestConfirmPasswordReset_UpdatesPassword(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	store := &fakeAuthStore{now: now}
	store.users = map[string]db.GetUserByEmailRow{
		"a@example.com": {
			ID:           pgUUID(uid),
			Email:        "a@example.com",
			PasswordHash: mustHash(),
			CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			DisplayName:  "Test User",
			VerifiedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		},
	}
	store.usersByID = map[uuid.UUID]db.GetUserByEmailRow{uid: store.users["a@example.com"]}
	store.resetTokens = map[string]db.PasswordResetToken{
		"tok": {
			ID:     pgUUID(uuid.New()),
			UserID: pgUUID(uid),
		},
	}

	s := NewAuthService(store)
	if err := s.ConfirmPasswordReset(context.Background(), "tok", "newpass123"); err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	u := store.usersByID[uid]
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("newpass123")) != nil {
		t.Fatalf("password not updated")
	}
	if len(store.resetTokens) != 0 {
		t.Fatalf("reset tokens should be cleared")
	}
}
