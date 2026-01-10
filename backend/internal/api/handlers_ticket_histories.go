package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hiromichi-5/forma/backend/internal/auth"
	"github.com/hiromichi-5/forma/backend/internal/service"
)

type ticketHistoriesService interface {
	ListTicketHistories(ctx context.Context, ticketID string, actor uuid.UUID) ([]service.TicketHistoryView, error)
}

type TicketHistoriesHandler struct{ Svc ticketHistoriesService }

func (h *TicketHistoriesHandler) GetV1TicketsTicketIdHistories(c *gin.Context, ticketID string) {
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

	histories, err := h.Svc.ListTicketHistories(c, ticketID, uid)
	if err != nil {
		switch err {
		case service.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR"})
		case service.ErrForbidden:
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"histories": histories})
}
