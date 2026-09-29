package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		in usecase.ListTicketsInput,
	) (usecase.TicketPage, error)
	CountTicketsByStatus(
		ctx context.Context,
		formID, userID uuid.UUID,
		emailQuery *string,
	) ([]usecase.TicketStatusCount, error)
	GetTicket(ctx context.Context, ticketID, userID uuid.UUID) (usecase.TicketDetail, error)
	UpdateTicket(
		ctx context.Context,
		ticketID, userID uuid.UUID,
		in usecase.UpdateTicketInput,
	) (usecase.TicketDetail, []usecase.NotificationResult, error)
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

	formID, err := uuid.Parse(c.Query("form_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	in := usecase.ListTicketsInput{EmailQuery: emailQuery(c)}
	for _, s := range c.QueryArray("status_id") {
		sid, err := uuid.Parse(s)
		if err != nil {
			handleError(c, entity.NewError(entity.CodeValidation))
			return
		}
		in.StatusIDs = append(in.StatusIDs, sid)
	}
	if s := c.Query("limit"); s != "" {
		limit, err := strconv.Atoi(s)
		if err != nil {
			handleError(c, entity.NewError(entity.CodeValidation))
			return
		}
		in.Limit = limit
	}
	if s := c.Query("cursor"); s != "" {
		cursor, err := decodeTicketCursor(s)
		if err != nil {
			handleError(c, entity.NewError(entity.CodeValidation))
			return
		}
		in.After = &cursor
	}

	page, err := h.uc.ListTickets(c, formID, userID, in)
	if err != nil {
		handleError(c, err)
		return
	}
	var next *string
	if page.NextCursor != nil {
		v := encodeTicketCursor(*page.NextCursor)
		next = &v
	}
	c.JSON(http.StatusOK, gin.H{
		"tickets":     toTicketSummaryListResp(page.Tickets),
		"next_cursor": next,
	})
}

func (h *TicketHandler) GetV1TicketsCounts(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	formID, err := uuid.Parse(c.Query("form_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	counts, err := h.uc.CountTicketsByStatus(c, formID, userID, emailQuery(c))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"counts": toTicketStatusCountListResp(counts)})
}

func emailQuery(c *gin.Context) *string {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return nil
	}
	return &q
}

type ticketCursorPayload struct {
	SubmittedAt time.Time `json:"submitted_at"`
	ID          uuid.UUID `json:"id"`
}

func encodeTicketCursor(cursor usecase.TicketCursor) string {
	b, _ := json.Marshal(ticketCursorPayload(cursor))
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeTicketCursor(s string) (usecase.TicketCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return usecase.TicketCursor{}, err
	}
	var p ticketCursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return usecase.TicketCursor{}, err
	}
	if p.SubmittedAt.IsZero() || p.ID == uuid.Nil {
		return usecase.TicketCursor{}, errors.New("incomplete ticket cursor")
	}
	return usecase.TicketCursor(p), nil
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

func (n nullableUUIDPayload) assigneeChange() entity.AssigneeChange {
	switch {
	case !n.set:
		return entity.KeepAssignee()
	case n.null:
		return entity.ClearAssignee()
	default:
		return entity.SetAssignee(n.value)
	}
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

	detail, results, err := h.uc.UpdateTicket(c, ticketID, userID, usecase.UpdateTicketInput{
		StatusID: statusID,
		Assignee: req.Assignee.assigneeChange(),
		Priority: (*entity.Priority)(req.Priority),
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTicketUpdateResp(detail, results))
}
