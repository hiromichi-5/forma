package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func (h *FormsHandler) GetV1Responses(c *gin.Context) {
	formID := c.Query("form_id")
	sinceStr := c.Query("since")

	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var since *time.Time
	if sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &t
		} else {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
			return
		}
	}
	rs, err := h.S.ListResponses(c, formID, since, uid)
	if err != nil {
		if err == service.ErrForbidden {
			c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		} else {
			c.JSON(500, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(200, gin.H{"responses": rs})
}
