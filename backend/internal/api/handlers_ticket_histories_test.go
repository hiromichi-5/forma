package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type fakeTicketHistoriesService struct {
	err       error
	histories []service.TicketHistoryView
}

func (f *fakeTicketHistoriesService) ListTicketHistories(_ context.Context, _ string, _ uuid.UUID) ([]service.TicketHistoryView, error) {
	return f.histories, f.err
}

func ticketHistoriesRouter(handler *TicketHistoriesHandler, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000001")
			c.Next()
		})
	}
	r.GET("/v1/tickets/:ticket_id/histories", func(c *gin.Context) {
		handler.GetV1TicketsTicketIdHistories(c, c.Param("ticket_id"))
	})
	return r
}

func TestTicketHistories_Get_Success(t *testing.T) {
	svc := &fakeTicketHistoriesService{histories: []service.TicketHistoryView{{FieldName: "status"}}}
	h := &TicketHistoriesHandler{Svc: svc}
	r := ticketHistoriesRouter(h, true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/tickets/00000000-0000-0000-0000-000000000010/histories", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestTicketHistories_Get_Unauthorized(t *testing.T) {
	h := &TicketHistoriesHandler{Svc: &fakeTicketHistoriesService{}}
	r := ticketHistoriesRouter(h, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/tickets/00000000-0000-0000-0000-000000000010/histories", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestTicketHistories_Get_NotFound(t *testing.T) {
	h := &TicketHistoriesHandler{Svc: &fakeTicketHistoriesService{err: service.ErrForbidden}}
	r := ticketHistoriesRouter(h, true)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/tickets/00000000-0000-0000-0000-000000000010/histories", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestTicketHistories_Get_Validation(t *testing.T) {
	h := &TicketHistoriesHandler{Svc: &fakeTicketHistoriesService{err: service.ErrValidation}}
	r := ticketHistoriesRouter(h, true)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/tickets/bad/histories", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
