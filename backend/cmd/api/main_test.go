package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
