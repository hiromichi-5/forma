package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/logger"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		reqLogger := slog.Default().With(
			slog.String("request_id", requestID),
		)

		ctx := logger.WithLogger(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", requestID)

		c.Next()

		if c.Request.URL.Path == "/healthz" {
			return
		}

		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.String("http.request.method", c.Request.Method),
			slog.String("url.path", c.Request.URL.Path),
			slog.Int("http.response.status_code", status),
			slog.String("http.server.request.duration", duration.String()),
			slog.String("client.address", c.ClientIP()),
		}

		if uid, ok := UserID(c); ok {
			attrs = append(attrs, slog.String("user_id", uid.String()))
		}

		if status >= 500 {
			reqLogger.Error("request completed", attrs...)
		} else {
			reqLogger.Info("request completed", attrs...)
		}
	}
}
