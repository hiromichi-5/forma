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

type fakeProfileStore struct {
	users map[string]db.GetUserByIDRow
}

func (f *fakeProfileStore) GetUserByID(_ context.Context, id pgtype.UUID) (db.GetUserByIDRow, error) {
	uid := uuid.UUID(id.Bytes).String()
	u, ok := f.users[uid]
	if !ok {
		return db.GetUserByIDRow{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeProfileStore) UpdateUserDisplayName(_ context.Context, arg db.UpdateUserDisplayNameParams) (db.UpdateUserDisplayNameRow, error) {
	uid := uuid.UUID(arg.ID.Bytes).String()
	u, ok := f.users[uid]
	if !ok {
		return db.UpdateUserDisplayNameRow{}, pgx.ErrNoRows
	}
	u.DisplayName = arg.DisplayName
	f.users[uid] = u
	return db.UpdateUserDisplayNameRow{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		DisplayName:  u.DisplayName,
		VerifiedAt:   u.VerifiedAt,
	}, nil
}

func (f *fakeProfileStore) DeleteUser(_ context.Context, id pgtype.UUID) (int64, error) {
	uid := uuid.UUID(id.Bytes).String()
	if _, ok := f.users[uid]; !ok {
		return 0, pgx.ErrNoRows
	}
	delete(f.users, uid)
	return 1, nil
}

func (f *fakeProfileStore) UpdateUserPasswordHash(_ context.Context, arg db.UpdateUserPasswordHashParams) error {
	uid := uuid.UUID(arg.ID.Bytes).String()
	u, ok := f.users[uid]
	if !ok {
		return pgx.ErrNoRows
	}
	u.PasswordHash = arg.PasswordHash
	f.users[uid] = u
	return nil
}

func fakeProfileStoreWith(userID, email, displayName string) *fakeProfileStore {
	return &fakeProfileStore{
		users: map[string]db.GetUserByIDRow{
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

func fakeProfileStoreWithPassword(userID, email, displayName, passwordHash string) *fakeProfileStore {
	store := fakeProfileStoreWith(userID, email, displayName)
	u := store.users[userID]
	u.PasswordHash = passwordHash
	store.users[userID] = u
	return store
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
	store := &fakeProfileStore{users: map[string]db.GetUserByIDRow{}}
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

func TestDeleteProfile_AlreadyDeleted(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	store := fakeProfileStoreWith(userID, "test@example.com", "Test User")
	s := NewProfileService(store)

	if err := s.DeleteProfile(context.Background(), userID); err != nil {
		t.Fatalf("first delete: want nil err, got %v", err)
	}

	err := s.DeleteProfile(context.Background(), userID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("second delete: want ErrUserNotFound, got %v", err)
	}
}

func TestChangePassword_Success(t *testing.T) {
	t.Helper()
	userID := "00000000-0000-0000-0000-000000000001"
	originalPassword := "oldpassword"
	hashed, err := bcrypt.GenerateFromPassword([]byte(originalPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt generate: %v", err)
	}
	store := fakeProfileStoreWithPassword(userID, "test@example.com", "Test User", string(hashed))
	s := NewProfileService(store)

	newPassword := "newpassword"
	if err := s.ChangePassword(context.Background(), userID, originalPassword, newPassword); err != nil {
		t.Fatalf("want nil err, got %v", err)
	}

	updated := store.users[userID]
	if bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte(newPassword)) != nil {
		t.Fatalf("stored password does not match new password")
	}
}

func TestChangePassword_IncorrectCurrentPassword(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	hashed, err := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt generate: %v", err)
	}
	store := fakeProfileStoreWithPassword(userID, "test@example.com", "Test User", string(hashed))
	s := NewProfileService(store)

	err = s.ChangePassword(context.Background(), userID, "wrong-pass", "newpassword")
	if !errors.Is(err, ErrIncorrectPassword) {
		t.Fatalf("want ErrIncorrectPassword, got %v", err)
	}
}

func TestChangePassword_ValidationErrors(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	hashed, err := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt generate: %v", err)
	}
	store := fakeProfileStoreWithPassword(userID, "test@example.com", "Test User", string(hashed))
	s := NewProfileService(store)

	if err := s.ChangePassword(context.Background(), userID, "", "newpassword"); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty current password: want ErrValidation, got %v", err)
	}
	if err := s.ChangePassword(context.Background(), userID, "correct-pass", "short"); !errors.Is(err, ErrValidation) {
		t.Fatalf("short new password: want ErrValidation, got %v", err)
	}
	if err := s.ChangePassword(context.Background(), "not-a-uuid", "correct-pass", "newpassword"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid user id: want ErrValidation, got %v", err)
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	store := &fakeProfileStore{users: map[string]db.GetUserByIDRow{}}
	s := NewProfileService(store)

	err := s.ChangePassword(context.Background(), "00000000-0000-0000-0000-000000000001", "password", "newpassword")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}
