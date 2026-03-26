package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func (h *FormsHandler) PostV1FormsFormIdSync(c *gin.Context, formID string) {
	uidStr, ok := auth.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
		return
	}
	synced, newTickets, last, err := h.S.SyncFormOnce(c, formID, uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrFormsNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	var lastStr string
	if !last.IsZero() {
		lastStr = last.UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, gin.H{
		"synced":     synced,
		"newTickets": newTickets,
		"last":       lastStr,
	})
}
