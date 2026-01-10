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

type fakeStatusesService struct {
	listErr    error
	createErr  error
	updateErr  error
	defaultErr error
	deleteErr  error
	statuses   []service.FormStatus
	status     service.FormStatus
}

func (f *fakeStatusesService) ListFormStatuses(_ context.Context, _ string, _ uuid.UUID) ([]service.FormStatus, error) {
	return f.statuses, f.listErr
}

func (f *fakeStatusesService) CreateFormStatus(_ context.Context, _ string, _ string, _ *string, _ int32, _ bool, _ uuid.UUID) (service.FormStatus, error) {
	return f.status, f.createErr
}

func (f *fakeStatusesService) UpdateFormStatus(_ context.Context, _ string, _ string, _ *string, _ *string, _ *int32, _ uuid.UUID) (service.FormStatus, error) {
	return f.status, f.updateErr
}

func (f *fakeStatusesService) SetDefaultFormStatus(_ context.Context, _ string, _ string, _ uuid.UUID) (service.FormStatus, error) {
	return f.status, f.defaultErr
}

func (f *fakeStatusesService) DeleteFormStatus(_ context.Context, _ string, _ string, _ uuid.UUID) error {
	return f.deleteErr
}

func statusesRouter(handler *StatusesHandler, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000001")
			c.Next()
		})
	}
	r.GET("/v1/forms/:form_id/statuses", func(c *gin.Context) {
		handler.GetV1FormsIdStatuses(c, c.Param("form_id"))
	})
	r.POST("/v1/forms/:form_id/statuses", func(c *gin.Context) {
		handler.PostV1FormsIdStatuses(c, c.Param("form_id"))
	})
	r.PATCH("/v1/forms/:form_id/statuses/:status_id", func(c *gin.Context) {
		handler.PatchV1FormsIdStatusesStatusId(c, c.Param("form_id"), c.Param("status_id"))
	})
	r.POST("/v1/forms/:form_id/statuses/:status_id/default", func(c *gin.Context) {
		handler.PostV1FormsIdStatusesStatusIdDefault(c, c.Param("form_id"), c.Param("status_id"))
	})
	r.DELETE("/v1/forms/:form_id/statuses/:status_id", func(c *gin.Context) {
		handler.DeleteV1FormsIdStatusesStatusId(c, c.Param("form_id"), c.Param("status_id"))
	})
	return r
}

func TestStatuses_Get_Success(t *testing.T) {
	svc := &fakeStatusesService{statuses: []service.FormStatus{{Name: "A", DisplayOrder: 1}}}
	h := &StatusesHandler{Svc: svc}
	r := statusesRouter(h, true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestStatuses_Get_Unauthorized(t *testing.T) {
	h := &StatusesHandler{Svc: &fakeStatusesService{}}
	r := statusesRouter(h, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestStatuses_Post_Success(t *testing.T) {
	svc := &fakeStatusesService{status: service.FormStatus{ID: uuid.New().String(), Name: "New", DisplayOrder: 1}}
	h := &StatusesHandler{Svc: svc}
	r := statusesRouter(h, true)
	body, _ := json.Marshal(map[string]any{"name": "New", "display_order": 1})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses", bytesReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestStatuses_Post_Validation(t *testing.T) {
	h := &StatusesHandler{Svc: &fakeStatusesService{}}
	r := statusesRouter(h, true)
	body, _ := json.Marshal(map[string]any{"name": ""})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses", bytesReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestStatuses_Patch_Success(t *testing.T) {
	svc := &fakeStatusesService{status: service.FormStatus{ID: uuid.New().String(), Name: "Edit", DisplayOrder: 2}}
	h := &StatusesHandler{Svc: svc}
	r := statusesRouter(h, true)
	body, _ := json.Marshal(map[string]any{"name": "Edit"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses/00000000-0000-0000-0000-000000000020", bytesReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestStatuses_Post_Default_Conflict(t *testing.T) {
	svc := &fakeStatusesService{defaultErr: service.ErrForbidden}
	h := &StatusesHandler{Svc: svc}
	r := statusesRouter(h, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses/00000000-0000-0000-0000-000000000020/default", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestStatuses_Delete_Conflict(t *testing.T) {
	svc := &fakeStatusesService{deleteErr: service.ErrConflict}
	h := &StatusesHandler{Svc: svc}
	r := statusesRouter(h, true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/forms/00000000-0000-0000-0000-000000000010/statuses/00000000-0000-0000-0000-000000000020", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
