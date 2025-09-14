package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (h *FormsHandler) GetV1Responses(c *gin.Context) {
	formID := c.Query("form_id")
	sinceStr := c.Query("since")
	var since *time.Time
	if sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &t
		} else {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
			return
		}
	}
	rs, err := h.S.ListResponses(c, formID, since)
	if err != nil {
		c.JSON(500, gin.H{"code": "INTERNAL"})
		return
	}
	c.JSON(200, gin.H{"responses": rs})
}
