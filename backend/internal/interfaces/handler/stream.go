package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
	"github.com/hiromichi-5/forma/backend/internal/usecase"
)

type StreamUseCase interface {
	CheckFormAccess(ctx context.Context, formID, userID uuid.UUID) error
	GetTicket(ctx context.Context, ticketID, userID uuid.UUID) (usecase.TicketDetail, error)
}

type EventSubscriber interface {
	Subscribe(formID uuid.UUID) (<-chan usecase.TicketEvent, func())
}

type StreamHandler struct {
	uc  StreamUseCase
	hub EventSubscriber
}

func NewStreamHandler(uc StreamUseCase, hub EventSubscriber) *StreamHandler {
	return &StreamHandler{uc: uc, hub: hub}
}

type ticketUpdatedEventData struct {
	TicketID string           `json:"ticket_id"`
	FormID   string           `json:"form_id"`
	Ticket   ticketDetailResp `json:"ticket"`
}

func (h *StreamHandler) GetV1FormsFormIdStream(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		handleError(c, entity.NewError(entity.CodeInvalidSession))
		return
	}

	formID, err := uuid.Parse(c.Param("form_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.CheckFormAccess(c.Request.Context(), formID, userID); err != nil {
		handleError(c, err)
		return
	}

	ch, unsubscribe := h.hub.Subscribe(formID)
	defer unsubscribe()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			detail, err := h.uc.GetTicket(ctx, event.TicketID, userID)
			if err != nil {
				continue
			}
			data := ticketUpdatedEventData{
				TicketID: event.TicketID.String(),
				FormID:   event.FormID.String(),
				Ticket:   toTicketDetailResp(detail),
			}
			b, err := json.Marshal(data)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: ticket_updated\ndata: %s\n\n", b)
			c.Writer.Flush()
		case <-ticker.C:
			fmt.Fprintf(c.Writer, "event: ping\ndata: {}\n\n")
			c.Writer.Flush()
		}
	}
}
