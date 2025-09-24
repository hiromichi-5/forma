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
)

type fakeProfileStore struct {
	users map[string]db.User
}

func (f *fakeProfileStore) GetUser(_ context.Context, id pgtype.UUID) (db.User, error) {
	uid := uuid.UUID(id.Bytes).String()
	u, ok := f.users[uid]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeProfileStore) UpdateUserDisplayName(_ context.Context, arg db.UpdateUserDisplayNameParams) (db.User, error) {
	uid := uuid.UUID(arg.ID.Bytes).String()
	u, ok := f.users[uid]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	u.DisplayName = arg.DisplayName
	f.users[uid] = u
	return u, nil
}

func (f *fakeProfileStore) DeleteUser(_ context.Context, id pgtype.UUID) error {
	uid := uuid.UUID(id.Bytes).String()
	if _, ok := f.users[uid]; !ok {
		return pgx.ErrNoRows
	}
	delete(f.users, uid)
	return nil
}

func fakeProfileStoreWith(userID, email, displayName string) *fakeProfileStore {
	return &fakeProfileStore{
		users: map[string]db.User{
			userID: {
				ID: pgtype.UUID{
					Bytes: uuid.MustParse(userID),
					Valid: true,
				},
				Email:        email,
				DisplayName:  displayName,
				PasswordHash: "hash",
				CreatedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
			},
		},
	}
}

func TestGetProfile_Success(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	store := fakeProfileStoreWith(userID, "test@example.com", "Test User")
	s := NewProfileService(store)

	user, err := s.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("want email test@example.com, got %s", user.Email)
	}
	if user.DisplayName != "Test User" {
		t.Fatalf("want display name 'Test User', got %s", user.DisplayName)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	store := &fakeProfileStore{users: map[string]db.User{}}
	s := NewProfileService(store)

	_, err := s.GetProfile(context.Background(), "00000000-0000-0000-0000-000000000001")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}

func TestUpdateDisplayName_Success(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	store := fakeProfileStoreWith(userID, "test@example.com", "Old Name")
	s := NewProfileService(store)

	user, err := s.UpdateDisplayName(context.Background(), userID, "New Name")
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}
	if user.DisplayName != "New Name" {
		t.Fatalf("want display name 'New Name', got %s", user.DisplayName)
	}
}

func TestUpdateDisplayName_EmptyName(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	store := fakeProfileStoreWith(userID, "test@example.com", "Old Name")
	s := NewProfileService(store)

	_, err := s.UpdateDisplayName(context.Background(), userID, "")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want ErrValidation, got %v", err)
	}
}

func TestDeleteProfile_Success(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	store := fakeProfileStoreWith(userID, "test@example.com", "Test User")
	s := NewProfileService(store)

	err := s.DeleteProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("want nil err, got %v", err)
	}

	// 削除されているかチェック
	_, err = s.GetProfile(context.Background(), userID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound after deletion, got %v", err)
	}
}
