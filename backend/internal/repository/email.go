package repository

import "context"

const (
	TemplateEmailVerification = "email-verification"
	TemplatePasswordReset     = "password-reset"
	TemplateInvite            = "invite"
)

type SendEmailInput struct {
	To           []string
	TemplateName string
	TemplateData map[string]string
}

type EmailSender interface {
	SendEmail(ctx context.Context, input SendEmailInput) error
}
