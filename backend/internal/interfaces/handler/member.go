package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/interfaces/middleware"
)

type MemberUseCase interface {
	ListMembers(ctx context.Context, formID, userID uuid.UUID) ([]entity.Member, error)
	AddMember(ctx context.Context, formID, userID uuid.UUID, email, role string) error
	ChangeRole(ctx context.Context, formID, userID, targetUserID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, formID, userID, targetUserID uuid.UUID) error
}

type MemberHandler struct {
	uc MemberUseCase
}

func NewMemberHandler(uc MemberUseCase) *MemberHandler {
	return &MemberHandler{uc: uc}
}

type memberAddReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"required,oneof=admin editor"`
}

type memberRoleUpdateReq struct {
	Role string `json:"role" binding:"required,oneof=admin editor"`
}

func (h *MemberHandler) GetV1FormsFormIdMembers(c *gin.Context) {
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

	members, err := h.uc.ListMembers(c, formID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": toMemberListResp(members)})
}

func (h *MemberHandler) PostV1FormsFormIdMembers(c *gin.Context) {
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

	var req memberAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.AddMember(c, formID, userID, req.Email, req.Role); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *MemberHandler) PutV1FormsFormIdMembersUserId(c *gin.Context) {
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
	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	var req memberRoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.ChangeRole(c, formID, userID, targetUserID, req.Role); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MemberHandler) DeleteV1FormsFormIdMembersUserId(c *gin.Context) {
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
	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		handleError(c, entity.NewError(entity.CodeValidation))
		return
	}

	if err := h.uc.RemoveMember(c, formID, userID, targetUserID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
