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

type fakeMembersService struct {
	requireEditorErr error
	requireAdminErr  error
	listMembers      []service.Member
	listErr          error
	addErr           error
	changeErr        error
	removeErr        error
}

func (f *fakeMembersService) RequireEditor(_ context.Context, _ string, _ uuid.UUID) error {
	return f.requireEditorErr
}

func (f *fakeMembersService) RequireAdmin(_ context.Context, _ string, _ uuid.UUID) error {
	return f.requireAdminErr
}

func (f *fakeMembersService) ListMembers(_ context.Context, _ string) ([]service.Member, error) {
	return f.listMembers, f.listErr
}

func (f *fakeMembersService) AddMember(_ context.Context, _ string, _ string, _ string) error {
	return f.addErr
}

func (f *fakeMembersService) ChangeRole(_ context.Context, _ string, _ string, _ string) error {
	return f.changeErr
}

func (f *fakeMembersService) RemoveMember(_ context.Context, _ string, _ string) error {
	return f.removeErr
}

func membersRouter(handler *MembersHandler, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000001")
			c.Next()
		})
	}
	r.GET("/v1/forms/:form_id/members", handler.GetV1FormsFormIdMembers)
	r.POST("/v1/forms/:form_id/members", handler.PostV1FormsFormIdMembers)
	r.PUT("/v1/forms/:form_id/members/:user_id", handler.PutV1FormsFormIdMembersUserId)
	r.DELETE("/v1/forms/:form_id/members/:user_id", handler.DeleteV1FormsFormIdMembersUserId)
	return r
}

func TestMembers_Get_Success(t *testing.T) {
	svc := &fakeMembersService{
		listMembers: []service.Member{{Email: "a@example.com", Role: "admin"}},
	}
	h := &MembersHandler{Svc: svc}
	r := membersRouter(h, true)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members",
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestMembers_Get_Unauthorized(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(
		w,
		httptest.NewRequest(
			http.MethodGet,
			"/v1/forms/00000000-0000-0000-0000-000000000010/members",
			nil,
		),
	)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestMembers_Get_NotFound(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{requireEditorErr: service.ErrForbidden}}
	r := membersRouter(h, true)
	w := httptest.NewRecorder()
	r.ServeHTTP(
		w,
		httptest.NewRequest(
			http.MethodGet,
			"/v1/forms/00000000-0000-0000-0000-000000000010/members",
			nil,
		),
	)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestMembers_Post_Success(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, true)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "role": "admin"})
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestMembers_Post_Validation(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, true)
	body, _ := json.Marshal(map[string]string{"email": "bad", "role": "admin"})
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestMembers_Post_NotFound(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{addErr: service.ErrUserNotFound}}
	r := membersRouter(h, true)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "role": "admin"})
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestMembers_Put_Validation(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, true)
	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members/bad",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestMembers_Put_Success(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, true)
	body, _ := json.Marshal(map[string]string{"role": "editor"})
	req := httptest.NewRequest(
		http.MethodPut,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members/00000000-0000-0000-0000-000000000002",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func TestMembers_Delete_Success(t *testing.T) {
	h := &MembersHandler{Svc: &fakeMembersService{}}
	r := membersRouter(h, true)
	req := httptest.NewRequest(
		http.MethodDelete,
		"/v1/forms/00000000-0000-0000-0000-000000000010/members/00000000-0000-0000-0000-000000000002",
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}
