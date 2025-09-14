package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func (h *FormsHandler) PostV1FormsFormIdSync(c *gin.Context, formID string) {
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	synced, newTickets, last, err := h.S.SyncFormOnce(c, formID, uid)
	if err != nil {
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN"})
		} else {
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
