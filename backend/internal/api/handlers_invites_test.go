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
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/hiromichi-5/forma/backend/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeInvitesService struct {
	createInvite db.FormInvite
	createErr    error
	listInvites  []db.FormInvite
	listErr      error
	deleteErr    error
	acceptErr    error
}

func (f *fakeInvitesService) CreateInvite(_ context.Context, _ string, _ string, _ string, _ uuid.UUID) (db.FormInvite, error) {
	return f.createInvite, f.createErr
}

func (f *fakeInvitesService) ListInvites(_ context.Context, _ string, _ uuid.UUID) ([]db.FormInvite, error) {
	return f.listInvites, f.listErr
}

func (f *fakeInvitesService) DeleteInvite(_ context.Context, _ string, _ string, _ uuid.UUID) error {
	return f.deleteErr
}

func (f *fakeInvitesService) AcceptInvite(_ context.Context, _ string, _ uuid.UUID) error {
	return f.acceptErr
}

func invitesRouter(handler *InvitesHandler, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000001")
			c.Next()
		})
	}
	r.POST("/v1/forms/:form_id/invites", handler.PostV1FormsFormIdInvites)
	r.GET("/v1/forms/:form_id/invites", handler.GetV1FormsFormIdInvites)
	r.DELETE("/v1/forms/:form_id/invites/:invite_id", handler.DeleteV1FormsFormIdInvitesInviteId)
	r.POST("/v1/invites/:invite_id/accept", func(c *gin.Context) {
		handler.PostV1InvitesInviteIdAccept(c, c.Param("invite_id"))
	})
	return r
}

func TestInvites_Create_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	invID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	svc := &fakeInvitesService{createInvite: db.FormInvite{
		ID:        pgtype.UUID{Bytes: invID, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: now, Valid: true},
	}}
	h := &InvitesHandler{Svc: svc}
	r := invitesRouter(h, true)

	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "role": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/v1/forms/00000000-0000-0000-0000-000000000020/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestInvites_Create_Unauthorized(t *testing.T) {
	h := &InvitesHandler{Svc: &fakeInvitesService{}}
	r := invitesRouter(h, false)
	body, _ := json.Marshal(map[string]string{"email": "a@example.com", "role": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/v1/forms/00000000-0000-0000-0000-000000000020/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestInvites_List_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	invID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	inviter := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	svc := &fakeInvitesService{listInvites: []db.FormInvite{{
		ID:        pgtype.UUID{Bytes: invID, Valid: true},
		Email:     "a@example.com",
		Role:      "admin",
		InvitedBy: pgtype.UUID{Bytes: inviter, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}}}
	h := &InvitesHandler{Svc: svc}
	r := invitesRouter(h, true)

	req := httptest.NewRequest(http.MethodGet, "/v1/forms/00000000-0000-0000-0000-000000000020/invites", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestInvites_Delete_NotFound(t *testing.T) {
	svc := &fakeInvitesService{deleteErr: service.ErrInviteNotFound}
	h := &InvitesHandler{Svc: svc}
	r := invitesRouter(h, true)

	req := httptest.NewRequest(http.MethodDelete, "/v1/forms/00000000-0000-0000-0000-000000000020/invites/00000000-0000-0000-0000-000000000010", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestInvites_Accept_Success(t *testing.T) {
	svc := &fakeInvitesService{}
	h := &InvitesHandler{Svc: svc}
	r := invitesRouter(h, true)

	req := httptest.NewRequest(http.MethodPost, "/v1/invites/00000000-0000-0000-0000-000000000010/accept", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}
