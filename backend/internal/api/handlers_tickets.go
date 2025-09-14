package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func (h *FormsHandler) GetV1Tickets(c *gin.Context) {
	formID := c.Query("form_id")
	status := c.Query("status")

	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	ts, err := h.S.ListTickets(c, formID, status, uid)
	if err != nil {
		if err == service.ErrForbidden {
			c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		} else {
			c.JSON(500, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(200, gin.H{"tickets": ts})
}

func (h *FormsHandler) GetV1TicketsTicketId(c *gin.Context, ticketID string) {
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	t, err := h.S.GetTicket(c, ticketID, uid)
	if err != nil {
		if err == service.ErrValidation {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		} else if err == service.ErrForbidden {
			c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		} else {
			c.JSON(404, gin.H{"code": "NOT_FOUND"})
		}
		return
	}
	c.JSON(200, t)
}

type patchTicketReq struct {
	Status     *string    `json:"status" binding:"omitempty,oneof=new in_progress done"`
	AssigneeID *uuid.UUID `json:"assignee_id"`
}

func (h *FormsHandler) PatchV1TicketsTicketId(c *gin.Context, ticketID string) {
	uidStr, _ := auth.UserID(c)
	uid, _ := uuid.Parse(uidStr)
	if uid == uuid.Nil {
		c.JSON(401, gin.H{"code": "UNAUTHORIZED"})
		return
	}

	var req patchTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		return
	}
	t, err := h.S.UpdateTicket(c, ticketID, req.Status, req.AssigneeID, uid)
	if err != nil {
		if err == service.ErrValidation {
			c.JSON(400, gin.H{"code": "VALIDATION_ERROR"})
		} else if err == service.ErrForbidden {
			c.JSON(403, gin.H{"code": "FORBIDDEN", "message": "insufficient role"})
		} else {
			c.JSON(500, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(200, t)
}
