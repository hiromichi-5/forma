package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
	"github.com/hiromichi-5/forma/backend/internal/repository"
)

const ctxUserID = "userID"

func SessionMiddleware(userRepo repository.UserRepository, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookieName == "" {
			abortInvalidSession(c)
			return
		}
		cookie, err := c.Request.Cookie(cookieName)
		if err != nil || cookie.Value == "" {
			abortInvalidSession(c)
			return
		}
		sid, err := uuid.Parse(cookie.Value)
		if err != nil {
			abortInvalidSession(c)
			return
		}
		log := logger.From(c.Request.Context())

		session, err := userRepo.GetSessionByID(c, sid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				log.Debug("session not found")
				abortInvalidSession(c)
				return
			}
			log.Error("session lookup failed", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "INTERNAL",
			})
			return
		}

		enriched := log.With(slog.String("user_id", session.UserID.String()))
		ctx := logger.WithLogger(c.Request.Context(), enriched)
		c.Request = c.Request.WithContext(ctx)

		c.Set(ctxUserID, session.UserID)
		c.Next()
	}
}

func abortInvalidSession(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    string(entity.CodeInvalidSession),
		"message": "セッションが無効です",
	})
}

func UserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok && id != uuid.Nil
}
