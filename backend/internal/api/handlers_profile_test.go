package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type fakeProfileService struct {
	users map[string]db.User
	err   error
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

func profileRouter(h *ProfileHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock middleware that sets user ID (auth.CtxUserID = "userID" を使用)
	r.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Next()
	})

	r.GET("/v1/me", h.GetV1Me)
	r.PUT("/v1/me", h.PutV1Me)
	r.DELETE("/v1/me", h.DeleteV1Me)
	return r
}

func TestGetV1Me_Success(t *testing.T) {
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
}

func TestPutV1Me_Success(t *testing.T) {
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

	body, _ := json.Marshal(map[string]string{"display_name": "New Name"})
	req := httptest.NewRequest("PUT", "/v1/me", bytes.NewReader(body))
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
}

func TestPutV1Me_EmptyDisplayName(t *testing.T) {
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
	req := httptest.NewRequest("PUT", "/v1/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
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
