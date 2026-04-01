package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type TicketHistoryUseCase interface {
	ListTicketHistories(
		ctx context.Context,
		ticketID, userID uuid.UUID,
	) ([]entity.TicketHistory, error)
}

type TicketHistoryHandler struct {
	uc TicketHistoryUseCase
}

func NewTicketHistoryHandler(uc TicketHistoryUseCase) *TicketHistoryHandler {
	return &TicketHistoryHandler{uc: uc}
}

func (h *TicketHistoryHandler) GetV1TicketsTicketIdHistories(c *gin.Context) {
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

	histories, err := h.uc.ListTicketHistories(c, ticketID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"histories": toTicketHistoryListResp(histories)})
}
