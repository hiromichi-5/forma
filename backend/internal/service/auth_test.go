package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type fakeStore struct {
	users map[string]db.User
}

func (f *fakeStore) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	u, ok := f.users[email]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeStore) CreateUser(_ context.Context, arg db.CreateUserParams) (db.User, error) {
	if f.users == nil {
		f.users = map[string]db.User{}
	}
	if _, ok := f.users[arg.Email]; ok {
		return db.User{}, errors.New("duplicate")
	}
	u := db.User{
		ID:           pgUUID(arg.ID.Bytes),
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		DisplayName:  arg.DisplayName,
	}
	f.users[arg.Email] = u
	return u, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func fakeStoreWith(email, plain string) *fakeStore {
	return &fakeStore{
		users: map[string]db.User{
			email: {
				ID: pgtype.UUID{
					Bytes: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Valid: true,
				},
				Email:        email,
				PasswordHash: mustHash(plain),
				CreatedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
				DisplayName: "Test User",
			},
		},
	}
}

func mustHash(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestAuthenticate_Success(t *testing.T) {
	s := NewAuthService(fakeStoreWith("a@example.com", "pass123"))
	uid, err := s.Authenticate(context.Background(), "a@example.com", "pass123")
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if uid == "" {
		t.Fatalf("want non-empty userID")
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	s := NewAuthService(fakeStoreWith("a@example.com", "pass123"))
	_, err := s.Authenticate(context.Background(), "a@example.com", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_UnknownEmail(t *testing.T) {
	s := NewAuthService(fakeStoreWith("a@example.com", "pass123"))
	_, err := s.Authenticate(context.Background(), "none@example.com", "pass123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestSignup_Success(t *testing.T) {
	f := &fakeStore{users: map[string]db.User{}}
	s := NewAuthService(f)
	uid, err := s.Signup(context.Background(), "new@example.com", "pass12345", "Test User")
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if uid == "" {
		t.Fatalf("want non-empty userID")
	}
}

func TestSignup_Conflict(t *testing.T) {
	f := fakeStoreWith("dup@example.com", "x")
	s := NewAuthService(f)
	_, err := s.Signup(context.Background(), "dup@example.com", "pass12345", "Test User")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}
