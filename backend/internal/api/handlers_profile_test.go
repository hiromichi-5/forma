package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeProfileService struct {
	users     map[string]db.User
	passwords map[string]string
	err       error
}

func (f *fakeProfileService) GetProfile(_ context.Context, userID string) (db.User, error) {
	if f.err != nil {
		return db.User{}, f.err
	}
	u, ok := f.users[userID]
	if !ok {
		return db.User{}, service.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeProfileService) UpdateDisplayName(_ context.Context, userID, displayName string) (db.User, error) {
	if f.err != nil {
		return db.User{}, f.err
	}
	if displayName == "" {
		return db.User{}, service.ErrValidation
	}
	u, ok := f.users[userID]
	if !ok {
		return db.User{}, service.ErrUserNotFound
	}
	u.DisplayName = displayName
	f.users[userID] = u
	return u, nil
}

func (f *fakeProfileService) DeleteProfile(_ context.Context, userID string) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.users[userID]; !ok {
		return service.ErrUserNotFound
	}
	delete(f.users, userID)
	return nil
}

func (f *fakeProfileService) ChangePassword(_ context.Context, userID, currentPassword, newPassword string) error {
	if f.err != nil {
		return f.err
	}
	if len(newPassword) < 8 {
		return service.ErrValidation
	}
	expected, ok := f.passwords[userID]
	if !ok {
		return service.ErrUserNotFound
	}
	if expected != currentPassword {
		return service.ErrIncorrectPassword
	}
	f.passwords[userID] = newPassword
	return nil
}

func profileRouter(h *ProfileHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Next()
	})

	r.GET("/v1/me", h.GetV1Me)
	r.PATCH("/v1/me", h.PatchV1Me)
	r.DELETE("/v1/me", h.DeleteV1Me)
	r.PATCH("/v1/me/password", h.PatchV1MePassword)
	return r
}

func profileRouterUnauthorized(h *ProfileHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/me", h.GetV1Me)
	r.PATCH("/v1/me", h.PatchV1Me)
	r.DELETE("/v1/me", h.DeleteV1Me)
	r.PATCH("/v1/me/password", h.PatchV1MePassword)
	return r
}

func TestGetV1Me_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Test User",
				VerifiedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			},
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	req := httptest.NewRequest("GET", "/v1/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp userProfileResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Fatalf("want email test@example.com, got %s", resp.Email)
	}
	if resp.DisplayName != "Test User" {
		t.Fatalf("want display name 'Test User', got %s", resp.DisplayName)
	}
	if resp.VerifiedAt == nil {
		t.Fatalf("verified_at should not be nil")
	}
}

func TestPatchV1Me_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Old Name",
				VerifiedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			},
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	body, _ := json.Marshal(map[string]string{"display_name": "New Name"})
	req := httptest.NewRequest("PATCH", "/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp userProfileResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if resp.DisplayName != "New Name" {
		t.Fatalf("want display name 'New Name', got %s", resp.DisplayName)
	}
	if resp.VerifiedAt == nil {
		t.Fatalf("verified_at should not be nil")
	}
}

func TestPatchV1Me_EmptyDisplayName(t *testing.T) {
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Old Name",
			},
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	body, _ := json.Marshal(map[string]string{"display_name": ""})
	req := httptest.NewRequest("PATCH", "/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestGetV1Me_Unauthorized(t *testing.T) {
	h := &ProfileHandler{Svc: &fakeProfileService{}}
	r := profileRouterUnauthorized(h)
	req := httptest.NewRequest("GET", "/v1/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestPatchV1Me_Unauthorized(t *testing.T) {
	h := &ProfileHandler{Svc: &fakeProfileService{}}
	r := profileRouterUnauthorized(h)
	body, _ := json.Marshal(map[string]string{"display_name": "New Name"})
	req := httptest.NewRequest("PATCH", "/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestDeleteV1Me_Unauthorized(t *testing.T) {
	h := &ProfileHandler{Svc: &fakeProfileService{}}
	r := profileRouterUnauthorized(h)
	req := httptest.NewRequest("DELETE", "/v1/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestPatchV1MePassword_Unauthorized(t *testing.T) {
	h := &ProfileHandler{Svc: &fakeProfileService{}}
	r := profileRouterUnauthorized(h)
	body, _ := json.Marshal(map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword",
	})
	req := httptest.NewRequest("PATCH", "/v1/me/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
func TestDeleteV1Me_Success(t *testing.T) {
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	req := httptest.NewRequest("DELETE", "/v1/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestPatchV1MePassword_Success(t *testing.T) {
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
		},
		passwords: map[string]string{
			"test-user-id": "oldpassword",
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	body, _ := json.Marshal(map[string]string{
		"current_password": "oldpassword",
		"new_password":     "newpassword",
	})
	req := httptest.NewRequest("PATCH", "/v1/me/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d body=%s", w.Code, w.Body.String())
	}
	if svc.passwords["test-user-id"] != "newpassword" {
		t.Fatalf("password not updated")
	}
}

func TestPatchV1MePassword_ValidationError(t *testing.T) {
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
		},
		passwords: map[string]string{
			"test-user-id": "oldpassword",
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	body, _ := json.Marshal(map[string]string{
		"current_password": "",
		"new_password":     "newpassword",
	})
	req := httptest.NewRequest("PATCH", "/v1/me/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestPatchV1MePassword_IncorrectPassword(t *testing.T) {
	svc := &fakeProfileService{
		users: map[string]db.User{
			"test-user-id": {
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
		},
		passwords: map[string]string{
			"test-user-id": "oldpassword",
		},
	}
	h := &ProfileHandler{Svc: svc}
	r := profileRouter(h)

	body, _ := json.Marshal(map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword",
	})
	req := httptest.NewRequest("PATCH", "/v1/me/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", w.Code, w.Body.String())
	}
}
