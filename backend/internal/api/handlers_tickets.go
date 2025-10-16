package api

import (
	"bytes"
	"encoding/json"

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
	Status   *string             `json:"status" binding:"omitempty,oneof=new in_progress done"`
	Assignee nullableUUIDPayload `json:"assignee_id"`
	Priority *string             `json:"priority" binding:"omitempty,oneof=High Medium Low"`
}

type nullableUUIDPayload struct {
	set   bool
	null  bool
	value uuid.UUID
}

func (n *nullableUUIDPayload) UnmarshalJSON(data []byte) error {
	n.set = true
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		n.null = true
		n.value = uuid.UUID{}
		return nil
	}
	var s string
	if err := json.Unmarshal(trimmed, &s); err != nil {
		return err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	n.null = false
	n.value = id
	return nil
}

func (n nullableUUIDPayload) Provided() bool { return n.set }

func (n nullableUUIDPayload) IsNull() bool { return n.set && n.null }

func (n nullableUUIDPayload) UUID() uuid.UUID { return n.value }

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
	var assigneeID *uuid.UUID
	clearAssignee := false
	if req.Assignee.Provided() {
		if req.Assignee.IsNull() {
			clearAssignee = true
		} else {
			id := req.Assignee.UUID()
			assigneeID = &id
		}
	}

	t, err := h.S.UpdateTicket(c, ticketID, req.Status, assigneeID, clearAssignee, req.Priority, uid)
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
