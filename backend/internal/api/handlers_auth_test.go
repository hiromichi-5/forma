package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type fakeAuthService struct {
	loginSessionID string
	loginErr       error
	signupID       string
	signupErr      error
	logoutErr      error
	verifyErr      error
	resendErr      error
	resetErr       error
	confirmErr     error
}

func (f *fakeAuthService) Authenticate(_ context.Context, _, _ string) (string, error) {
	if f.loginErr != nil {
		return "", f.loginErr
	}
	return f.loginSessionID, nil
}

func (f *fakeAuthService) Signup(_ context.Context, _, _, _ string) (string, error) {
	if f.signupErr != nil {
		return "", f.signupErr
	}
	return f.signupID, nil
}

func (f *fakeAuthService) Logout(_ context.Context, _ string) error {
	return f.logoutErr
}

func (f *fakeAuthService) VerifyEmail(_ context.Context, _ string) error {
	return f.verifyErr
}

func (f *fakeAuthService) ResendEmailVerification(_ context.Context, _ string) error {
	return f.resendErr
}

func (f *fakeAuthService) RequestPasswordReset(_ context.Context, _ string) error {
	return f.resetErr
}

func (f *fakeAuthService) ConfirmPasswordReset(_ context.Context, _, _ string) error {
	return f.confirmErr
}

func router(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/auth/login", h.PostV1AuthLogin)
	r.POST("/v1/auth/signup", h.PostV1AuthSignup)
	r.POST("/v1/auth/logout", h.PostV1AuthLogout)
	r.POST("/v1/auth/verify-email", h.PostV1AuthVerifyEmail)
	r.POST("/v1/auth/verify-email/resend", h.PostV1AuthVerifyEmailResend)
	r.POST("/v1/auth/password-reset", h.PostV1AuthPasswordReset)
	r.POST("/v1/auth/password-reset/confirm", h.PostV1AuthPasswordResetConfirm)
	return r
}

func TestLogin_Success(t *testing.T) {
	sid := uuid.New().String()
	h := &AuthHandler{
		Svc:    &fakeAuthService{loginSessionID: sid},
		Cookie: AuthCookieConfig{Name: "forma_token"},
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
	cookie := findCookie(w.Result().Cookies(), "forma_token")
	if cookie == nil {
		t.Fatalf("want forma_token cookie")
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["session_id"] != sid {
		t.Fatalf("want session_id %s, got %s", sid, resp["session_id"])
	}
}

func TestLogin_EmailNotVerified(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{loginErr: service.ErrEmailNotVerified}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "pass123"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{loginErr: service.ErrInvalidCredentials}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "password": "wrong"})
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSignup_Success(t *testing.T) {
	uid := uuid.New().String()
	h := &AuthHandler{Svc: &fakeAuthService{signupID: uid}, Cookie: AuthCookieConfig{Name: "forma_token"}}
	r := router(h)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com", "password": "password123", "display_name": "TestUser"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", w.Code, w.Body.String())
	}
	if c := findCookie(w.Result().Cookies(), "forma_token"); c != nil {
		t.Fatalf("signup should not set cookie")
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["id"] != uid {
		t.Fatalf("want id %s, got %s", uid, resp["id"])
	}
}

func TestSignup_Conflict(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{signupErr: service.ErrConflict}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "password": "password123", "display_name": "TestUser"})
	req := httptest.NewRequest("POST", "/v1/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
}

func TestLogout_Unauthorized(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{}}
	r := router(h)
	req := httptest.NewRequest("POST", "/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	sid := uuid.New().String()
	h := &AuthHandler{Svc: &fakeAuthService{}, Cookie: AuthCookieConfig{Name: "forma_token"}}
	r := router(h)

	req := httptest.NewRequest("POST", "/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "forma_token", Value: sid})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
	logoutCookie := findCookie(w.Result().Cookies(), "forma_token")
	if logoutCookie == nil || logoutCookie.MaxAge != -1 {
		t.Fatalf("logout cookie missing or invalid")
	}
}

func TestVerifyEmail_NotFound(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{verifyErr: service.ErrTokenNotFound}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"token": "tok"})
	req := httptest.NewRequest("POST", "/v1/auth/verify-email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"token": "tok"})
	req := httptest.NewRequest("POST", "/v1/auth/verify-email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func TestResend_Verification(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com"})
	req := httptest.NewRequest("POST", "/v1/auth/verify-email/resend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d", w.Code)
	}
}

func TestPasswordReset_Request(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com"})
	req := httptest.NewRequest("POST", "/v1/auth/password-reset", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d", w.Code)
	}
}

func TestPasswordReset_Confirm_NotFound(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{confirmErr: service.ErrTokenNotFound}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"token": "bad", "new_password": "newpass123"})
	req := httptest.NewRequest("POST", "/v1/auth/password-reset/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestPasswordReset_Confirm_Success(t *testing.T) {
	h := &AuthHandler{Svc: &fakeAuthService{}}
	r := router(h)
	body, _ := json.Marshal(map[string]string{"token": "tok", "new_password": "newpass123"})
	req := httptest.NewRequest("POST", "/v1/auth/password-reset/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
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
