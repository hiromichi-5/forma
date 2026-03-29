package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestHealthz_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d", w.Code)
	}
	if got := w.Body.String(); got != "ok" {
		t.Fatalf("body: want %q got %q", "ok", got)
	}
}

type fakeAuth struct {
	email string
	pass  string
	sid   string
}

func (f *fakeAuth) Authenticate(_ context.Context, e, p string) (string, error) {
	if e == f.email && p == f.pass {
		return f.sid, nil
	}
	return "", nil
}

func (f *fakeAuth) Signup(_ context.Context, e, p, displayName string) (string, error) {
	if e != "" && p != "" && displayName != "" {
		return "dummy", nil
	}
	return "", nil
}

func (f *fakeAuth) Logout(_ context.Context, _ string) error { return nil }

func (f *fakeAuth) VerifyEmail(_ context.Context, _ string) error { return nil }

func (f *fakeAuth) ResendEmailVerification(_ context.Context, _ string) error { return nil }

func (f *fakeAuth) RequestPasswordReset(_ context.Context, _ string) error { return nil }

func (f *fakeAuth) ConfirmPasswordReset(_ context.Context, _, _ string) error { return nil }

type fakeSessionStore struct {
	sessions map[uuid.UUID]db.Session
}

func (f *fakeSessionStore) GetSessionByID(_ context.Context, id pgtype.UUID) (db.Session, error) {
	if f.sessions == nil {
		return db.Session{}, pgx.ErrNoRows
	}
	s, ok := f.sessions[id.Bytes]
	if !ok {
		return db.Session{}, pgx.ErrNoRows
	}
	return s, nil
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := NewRouter()
	h := &api.AuthHandler{
		Svc:    &fakeAuth{"a@example.com", "pass123", "session-1"},
		Cookie: api.AuthCookieConfig{Name: "forma_token"},
	}
	r.POST("/v1/auth/login", h.PostV1AuthLogin)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d body=%s", w.Code, w.Body.String())
	}
	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie missing")
	}
	if cookie.Value != "session-1" {
		t.Fatalf("unexpected session id: %s", cookie.Value)
	}
}

func TestWhoAmI_AuthFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := NewRouter()
	cookieCfg := api.AuthCookieConfig{Name: "forma_token"}
	h := &api.AuthHandler{
		Svc:    &fakeAuth{"a@example.com", "pass123", "session-42"},
		Cookie: cookieCfg,
	}

	r.POST("/v1/auth/login", h.PostV1AuthLogin)
	authz := r.Group("/v1")
	store := &fakeSessionStore{}
	sid, _ := uuid.Parse("00000000-0000-0000-0000-000000000042")
	uid, _ := uuid.Parse("00000000-0000-0000-0000-000000000043")
	store.sessions = map[uuid.UUID]db.Session{
		sid: {
			ID:     pgtype.UUID{Bytes: sid, Valid: true},
			UserID: pgtype.UUID{Bytes: uid, Valid: true},
		},
	}
	authz.Use(auth.SessionMiddleware(store, cookieCfg.Name))
	authz.GET("/whoami", func(c *gin.Context) {
		if uid, ok := auth.UserID(c); ok {
			c.JSON(http.StatusOK, gin.H{"user_id": uid})
			return
		}
		c.Status(http.StatusInternalServerError)
	})

	req0 := httptest.NewRequest(http.MethodGet, "/v1/whoami", nil)
	w0 := httptest.NewRecorder()
	r.ServeHTTP(w0, req0)
	if w0.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", w0.Code)
	}

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w1.Code, w1.Body.String())
	}
	cookie := findCookie(w1.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie missing")
	}
	cookie.Value = sid.String()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/whoami", nil)
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("whoami status: want 200 got %d body=%s", w2.Code, w2.Body.String())
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}
