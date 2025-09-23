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
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type fakeAuth struct {
	email string
	pass  string
	uid   string
	err   error
}

func (f *fakeAuth) Authenticate(_ context.Context, e, p string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if e == f.email && p == f.pass {
		return f.uid, nil
	}
	return "", service.ErrInvalidCredentials
}

func (f *fakeAuth) Signup(_ context.Context, e, p string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if e == "" || p == "" {
		return "", service.ErrValidation
	}
	if e == f.email {
		return "", service.ErrConflict
	}
	return "new-user-id", nil
}

func router(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/auth/login", h.PostV1AuthLogin)
	r.POST("/v1/auth/signup", h.PostV1AuthSignup)
	return r
}

func TestLogin_Success(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := auth.Signer{Secret: []byte("k"), TTL: time.Hour, Now: func() time.Time { return base }}

	h := &AuthHandler{
		Svc: &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT: s,
	}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("bad json: %v %q", err, w.Body.String())
	}

	claims, err := s.Parse(resp.Token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Subject != "u-1" {
		t.Fatalf("want sub u-1, got %s", claims.Subject)
	}
	if claims.IssuedAt.Time != base || claims.ExpiresAt.Time != base.Add(time.Hour) {
		t.Fatalf("claims time mismatch")
	}
}

func TestSignup_Success(t *testing.T) {
	s := auth.Signer{Secret: []byte("k"), TTL: time.Hour}
	h := &AuthHandler{Svc: &fakeAuth{email: "exists@example.com"}, JWT: s}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("bad json: %v %q", err, w.Body.String())
	}
}

func TestSignup_Conflict(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuth{email: "dup@example.com"}, JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "password": "password123"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSignup_Validation(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuth{}, JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour}}
	r := router(h)
	// メール不正
	body, _ := json.Marshal(map[string]string{"email": "bad", "password": "password123"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := &AuthHandler{
		Svc: &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour},
	}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "wrong"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestLogin_ValidationError(t *testing.T) {
	h := &AuthHandler{
		Svc: &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour},
	}
	r := router(h)

	// email欠落
	body, _ := json.Marshal(map[string]string{"password": "x"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}

	// emailフォーマット不正
	body2, _ := json.Marshal(map[string]string{"email": "bad", "password": "x"})
	req2 := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w2.Code)
	}
}
