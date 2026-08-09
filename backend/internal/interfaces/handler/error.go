package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hiromichi-5/forma/backend/internal/entity"
	"github.com/hiromichi-5/forma/backend/internal/logger"
)

type errorDef struct {
	Status  int
	Message string
}

var errorDefs = map[entity.Code]errorDef{
	entity.CodeInvalidCredentials:        {http.StatusUnauthorized, "メールアドレスまたはパスワードが正しくありません"},
	entity.CodeInvalidSession:            {http.StatusUnauthorized, "セッションが無効です"},
	entity.CodeEmailNotVerified:          {http.StatusForbidden, "メールの認証が完了していません"},
	entity.CodeForbidden:                 {http.StatusForbidden, "この操作を実行する権限がありません"},
	entity.CodeResourceHidden:            {http.StatusNotFound, "リソースが見つかりません"},
	entity.CodeUserNotFound:              {http.StatusNotFound, "ユーザーが見つかりません"},
	entity.CodeFormNotFound:              {http.StatusNotFound, "フォームが見つかりません"},
	entity.CodeFormNotShared:             {http.StatusNotFound, "フォームが見つかりません"},
	entity.CodeTokenNotFound:             {http.StatusNotFound, "トークンが見つかりません"},
	entity.CodeInviteNotFound:            {http.StatusNotFound, "招待が見つかりません"},
	entity.CodeInviteExpired:             {http.StatusNotFound, "招待の有効期限が切れています"},
	entity.CodeAlreadyMember:             {http.StatusConflict, "既にメンバーです"},
	entity.CodeIncorrectPassword:         {http.StatusForbidden, "現在のパスワードが正しくありません"},
	entity.CodeLastAdmin:                 {http.StatusConflict, "管理者は最低1名必要です"},
	entity.CodeConflict:                  {http.StatusConflict, "リソースが競合しています"},
	entity.CodeFormAlreadyRegistered:     {http.StatusConflict, "このフォームは既に登録されています"},
	entity.CodeActiveInviteAlreadyExists: {http.StatusConflict, "このメールアドレスへの有効な招待が既に存在します"},
	entity.CodeStatusConflict:            {http.StatusConflict, "ステータス名または表示順が競合しています"},
	entity.CodeNotificationDisabled:      {http.StatusConflict, "管理者によってこの通知は無効化されています"},
	entity.CodeRespondentEmailMissing: {
		http.StatusConflict,
		"回答者のメールアドレスが登録されていません",
	},
	entity.CodeNotificationRateLimited: {
		http.StatusTooManyRequests,
		"メールの送信間隔には制限があります。しばらく時間を置いてから再度実行してください",
	},
	entity.CodeValidation: {http.StatusBadRequest, "入力内容に誤りがあります"},
}

type errorResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message,omitempty"`
	Fields  []fieldErrorResp `json:"fields,omitempty"`
}

type fieldErrorResp struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

func handleError(c *gin.Context, err error) {
	log := logger.From(c.Request.Context())

	var domainErr *entity.Error
	if !errors.As(err, &domainErr) {
		log.Error("unexpected error", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Code: "INTERNAL"})
		return
	}

	def, ok := errorDefs[domainErr.Code]
	if !ok {
		log.Error("unmapped domain error", "code", domainErr.Code, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Code: "INTERNAL"})
		return
	}

	log.Debug("domain error", "code", domainErr.Code, "status", def.Status)

	resp := errorResponse{
		Code:    string(domainErr.Code),
		Message: def.Message,
	}
	if len(domainErr.Fields) > 0 {
		resp.Fields = make([]fieldErrorResp, len(domainErr.Fields))
		for i, f := range domainErr.Fields {
			resp.Fields[i] = fieldErrorResp{
				Field: f.Field,
				Code:  string(f.Code),
			}
		}
	}
	c.JSON(def.Status, resp)
}
