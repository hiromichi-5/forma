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

func (f *fakeAuth) Signup(_ context.Context, e, p, displayName string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if e == "" || p == "" || displayName == "" {
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
	r.POST("/v1/auth/logout", h.PostV1AuthLogout)
	return r
}

func TestLogin_Success(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := auth.Signer{Secret: []byte("k"), TTL: time.Hour, Now: func() time.Time { return base }}

	h := &AuthHandler{
		Svc:    &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT:    s,
		Cookie: AuthCookieConfig{Name: "forma_token"},
	}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d body=%s", w.Code, w.Body.String())
	}

	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("want forma_token cookie")
	}
	if !cookie.HttpOnly {
		t.Fatalf("cookie must be HttpOnly")
	}
	claims, err := s.Parse(cookie.Value)
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
	h := &AuthHandler{Svc: &fakeAuth{email: "exists@example.com"}, JWT: s, Cookie: AuthCookieConfig{Name: "forma_token"}}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com", "password": "password123", "display_name": "TestUser"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d body=%s", w.Code, w.Body.String())
	}
	if c := findCookie(w.Result().Cookies(), "forma_token"); c == nil {
		t.Fatalf("cookie not found")
	}
}

func TestSignup_Conflict(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuth{email: "dup@example.com"}, JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour}, Cookie: AuthCookieConfig{}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "password": "password123", "display_name": "TestUser"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSignup_Validation(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuth{}, JWT: auth.Signer{Secret: []byte("k"), TTL: time.Hour}, Cookie: AuthCookieConfig{}}
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
		Svc:    &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT:    auth.Signer{Secret: []byte("k"), TTL: time.Hour},
		Cookie: AuthCookieConfig{},
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
		Svc:    &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"},
		JWT:    auth.Signer{Secret: []byte("k"), TTL: time.Hour},
		Cookie: AuthCookieConfig{},
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

func TestLogout_ClearsCookie(t *testing.T) {
	s := auth.Signer{Secret: []byte("k"), TTL: time.Hour}
	h := &AuthHandler{Svc: &fakeAuth{email: "a@example.com", pass: "pass123", uid: "u-1"}, JWT: s, Cookie: AuthCookieConfig{Name: "forma_token"}}
	r := router(h)

	loginBody, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if c := findCookie(w.Result().Cookies(), "forma_token"); c == nil {
		t.Fatalf("cookie not issued")
	}

	logoutReq := httptest.NewRequest("POST", "/v1/auth/logout", nil)
	logoutW := httptest.NewRecorder()
	r.ServeHTTP(logoutW, logoutReq)
	if logoutW.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", logoutW.Code)
	}
	logoutCookie := findCookie(logoutW.Result().Cookies(), "forma_token")
	if logoutCookie == nil || logoutCookie.MaxAge != -1 {
		t.Fatalf("logout cookie missing or invalid")
	}
}

func TestSetAuthCookie_RoundsUpSubSecondTTL(t *testing.T) {
	h := &AuthHandler{
		Svc:    &fakeAuth{},
		JWT:    auth.Signer{Secret: []byte("k"), TTL: 500 * time.Millisecond},
		Cookie: AuthCookieConfig{Name: "forma_token"},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	h.setAuthCookie(c, "tok")

	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie not found")
	}
	if cookie.MaxAge != 1 {
		t.Fatalf("max-age should round up to 1, got %d", cookie.MaxAge)
	}
}

func TestSetAuthCookie_NegativeTTLDeletesCookie(t *testing.T) {
	h := &AuthHandler{
		Svc:    &fakeAuth{},
		JWT:    auth.Signer{Secret: []byte("k"), TTL: -time.Second},
		Cookie: AuthCookieConfig{Name: "forma_token"},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	h.setAuthCookie(c, "tok")

	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie not found")
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("max-age should be -1 for negative TTL, got %d", cookie.MaxAge)
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
