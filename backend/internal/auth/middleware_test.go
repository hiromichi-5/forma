package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeSessionStore struct {
	sessions map[uuid.UUID]db.Session
}

func (f *fakeSessionStore) GetSessionByID(_ context.Context, id pgtype.UUID) (db.Session, error) {
	if f.sessions == nil {
		return db.Session{}, pgx.ErrNoRows
	}
	s, ok := f.sessions[id.Bytes]
	if !ok {
		return db.Session{}, pgx.ErrNoRows
	}
	return s, nil
}

func routerWith(store SessionStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/p", SessionMiddleware(store, "forma_token"), func(c *gin.Context) {
		uid, _ := UserID(c)
		c.String(http.StatusOK, uid)
	})
	return r
}

func TestSession_MissingToken(t *testing.T) {
	r := routerWith(&fakeSessionStore{})
	req := httptest.NewRequest("GET", "/p", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSession_InvalidToken(t *testing.T) {
	r := routerWith(&fakeSessionStore{})
	req := httptest.NewRequest("GET", "/p", nil)
	req.AddCookie(&http.Cookie{Name: "forma_token", Value: "bad"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSession_NotFound(t *testing.T) {
	r := routerWith(&fakeSessionStore{})
	req := httptest.NewRequest("GET", "/p", nil)
	req.AddCookie(&http.Cookie{Name: "forma_token", Value: uuid.New().String()})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSession_OK(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sid := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	store := &fakeSessionStore{sessions: map[uuid.UUID]db.Session{
		sid: {
			ID:     pgtype.UUID{Bytes: sid, Valid: true},
			UserID: pgtype.UUID{Bytes: uid, Valid: true},
		},
	}}
	r := routerWith(store)
	req := httptest.NewRequest("GET", "/p", nil)
	req.AddCookie(&http.Cookie{Name: "forma_token", Value: sid.String()})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if w.Body.String() != uid.String() {
		t.Fatalf("want body %s, got %s", uid.String(), w.Body.String())
	}
}
