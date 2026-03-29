package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const CtxUserID = "userID"

type SessionStore interface {
	GetSessionByID(ctx context.Context, id pgtype.UUID) (db.Session, error)
}

func SessionMiddleware(store SessionStore, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookieName == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing token",
			})
			return
		}
		cookie, err := c.Request.Cookie(cookieName)
		if err != nil || cookie.Value == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing token",
			})
			return
		}
		sid, err := uuid.Parse(cookie.Value)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid token",
			})
			return
		}
		session, err := store.GetSessionByID(c, pgtype.UUID{Bytes: sid, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid token",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "INTERNAL",
			})
			return
		}
		c.Set(CtxUserID, uuid.UUID(session.UserID.Bytes).String())
		c.Next()
	}
}

func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}
