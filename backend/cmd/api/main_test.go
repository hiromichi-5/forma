package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hiromichi-5/forma/backend/internal/api"
	"github.com/hiromichi-5/forma/backend/internal/auth"
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
	uid   string
}

func (f *fakeAuth) Authenticate(_ context.Context, e, p string) (string, error) {
	if e == f.email && p == f.pass {
		return f.uid, nil
	}
	return "", nil
}

func (f *fakeAuth) Signup(_ context.Context, e, p, displayName string) (string, error) {
	if e != "" && p != "" && displayName != "" {
		return "dummy", nil
	}
	return "", nil
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := NewRouter()
	base := time.Unix(1_700_000_000, 0)
	signer := auth.Signer{Secret: []byte("k"), TTL: time.Hour, Now: func() time.Time { return base }}
	h := &api.AuthHandler{Svc: &fakeAuth{"a@example.com", "pass123", "u-1"}, JWT: signer, Cookie: api.AuthCookieConfig{Name: "forma_token"}}
	r.POST("/v1/auth/login", h.PostV1AuthLogin)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status: want 204 got %d body=%s", w.Code, w.Body.String())
	}
	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie missing")
	}
	claims, err := signer.Parse(cookie.Value)
	if err != nil || claims.Subject != "u-1" {
		t.Fatalf("claims invalid: %v sub=%v", err, claims.Subject)
	}
}

func TestWhoAmI_AuthFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := NewRouter()
	base := time.Unix(1_700_000_000, 0)
	signer := auth.Signer{Secret: []byte("k"), TTL: time.Hour, Now: func() time.Time { return base }}
	cookieCfg := api.AuthCookieConfig{Name: "forma_token"}
	h := &api.AuthHandler{Svc: &fakeAuth{"a@example.com", "pass123", "u-42"}, JWT: signer, Cookie: cookieCfg}

	r.POST("/v1/auth/login", h.PostV1AuthLogin)
	authz := r.Group("/v1")
	authz.Use(auth.BearerMiddleware(signer, cookieCfg.Name))
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
	if w1.Code != http.StatusNoContent {
		t.Fatalf("login failed: %d %s", w1.Code, w1.Body.String())
	}
	cookie := findCookie(w1.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("cookie missing")
	}
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
