package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type SyncUseCase interface {
	SyncFormOnce(
		ctx context.Context,
		formID, userID uuid.UUID,
	) (newTickets int, lastSync time.Time, err error)
}

type SyncHandler struct {
	uc SyncUseCase
}

func NewSyncHandler(uc SyncUseCase) *SyncHandler {
	return &SyncHandler{uc: uc}
}

func (h *SyncHandler) PostV1FormsFormIdSync(c *gin.Context) {
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

	newTickets, lastSync, err := h.uc.SyncFormOnce(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	var lastStr string
	if !lastSync.IsZero() {
		lastStr = lastSync.UTC().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, syncResp{
		Synced:     true,
		NewTickets: newTickets,
		Last:       lastStr,
	})
}
