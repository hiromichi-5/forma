package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// func fixed(t time.Time) func() time.Time { return func() time.Time { return t } }

func routerWith(s Signer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/p", BearerMiddleware(s), func(c *gin.Context) {
		uid, _ := UserID(c)
		c.String(http.StatusOK, uid)
	})
	return r
}

func TestBearer_MissingToken(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}
	r := routerWith(s)

	req := httptest.NewRequest("GET", "/p", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestBearer_InvalidSchemeOrEmpty(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}
	r := routerWith(s)

	req := httptest.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Basic abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}

	req2 := httptest.NewRequest("GET", "/p", nil)
	req2.Header.Set("Authorization", "Bearer ")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w2.Code)
	}
}

func TestBearer_InvalidToken(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}
	r := routerWith(s)

	req := httptest.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestBearer_ExpiredToken(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}

	tok, err := s.Issue("u-1")
	if err != nil {
		t.Fatalf("issue err: %v", err)
	}

	s.Now = fixed(base.Add(2 * time.Hour))
	r := routerWith(s)

	req := httptest.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 expired, got %d", w.Code)
	}
}

func TestBearer_OK(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	s := Signer{Secret: []byte("k"), TTL: time.Hour, Now: fixed(base)}
	tok, err := s.Issue("u-42")
	if err != nil {
		t.Fatalf("issue err: %v", err)
	}
	r := routerWith(s)

	req := httptest.NewRequest("GET", "/p", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if w.Body.String() != "u-42" {
		t.Fatalf("want body u-42, got %q", w.Body.String())
	}
}
