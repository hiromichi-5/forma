package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
)

type TicketUseCase interface {
	ListTickets(
		ctx context.Context,
		formID, userID uuid.UUID,
		statusID *uuid.UUID,
	) ([]usecase.TicketSummary, error)
	GetTicket(ctx context.Context, ticketID, userID uuid.UUID) (usecase.TicketDetail, error)
	UpdateTicket(
		ctx context.Context,
		ticketID, userID uuid.UUID,
		statusID *uuid.UUID,
		assigneeID *uuid.UUID,
		clearAssignee bool,
		priority *string,
	) (usecase.TicketDetail, error)
}

type TicketHandler struct {
	uc TicketUseCase
}

func NewTicketHandler(uc TicketUseCase) *TicketHandler {
	return &TicketHandler{uc: uc}
}

func (h *TicketHandler) GetV1Tickets(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	formIDStr := c.Query("form")
	formID, err := uuid.Parse(formIDStr)
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var statusID *uuid.UUID
	if s := c.Query("status_id"); s != "" {
		sid, err := uuid.Parse(s)
		if err != nil {
			handleError(c, entity.NewError(entity.CodeValidation))
			return
		}
		statusID = &sid
	}

	tickets, err := h.uc.ListTickets(c, formID, userID, statusID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": toTicketSummaryListResp(tickets)})
}

func (h *TicketHandler) GetV1TicketsTicketId(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	ticketID, err := uuid.Parse(c.Param("ticket_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	detail, err := h.uc.GetTicket(c, ticketID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTicketDetailResp(detail))
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

func (h *TicketHandler) PatchV1TicketsTicketId(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}
	ticketID, err := uuid.Parse(c.Param("ticket_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var req patchTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var statusID *uuid.UUID
	if req.StatusID != nil {
		sid, err := uuid.Parse(*req.StatusID)
		if err != nil {
			handleError(c, entity.NewError(entity.CodeValidation))
			return
		}
		statusID = &sid
	}

	var assigneeID *uuid.UUID
	clearAssignee := false
	if req.Assignee.set {
		if req.Assignee.null {
			clearAssignee = true
		} else {
			id := req.Assignee.value
			assigneeID = &id
		}
	}

	detail, err := h.uc.UpdateTicket(
		c,
		ticketID,
		userID,
		statusID,
		assigneeID,
		clearAssignee,
		req.Priority,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTicketDetailResp(detail))
}
