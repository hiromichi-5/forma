package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

func (h *FormsHandler) GetV1Tickets(c *gin.Context) {
	formID := c.Query("form")
	statusID := c.Query("status_id")

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

	ts, err := h.S.ListTickets(c, formID, statusID, uid)
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
	c.JSON(http.StatusOK, gin.H{"tickets": ts})
}

func (h *FormsHandler) GetV1TicketsTicketId(c *gin.Context, ticketID string) {
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

	t, err := h.S.GetTicket(c, ticketID, uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, t)
}

type patchTicketReq struct {
	StatusID *string             `json:"status_id"`
	Assignee nullableUUIDPayload `json:"assignee_id"`
	Priority *string             `json:"priority"    binding:"omitempty,oneof=high medium low"`
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

	var req patchTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
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

	t, err := h.S.UpdateTicket(
		c,
		ticketID,
		req.StatusID,
		assigneeID,
		clearAssignee,
		req.Priority,
		uid,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}
	c.JSON(http.StatusOK, t)
}
