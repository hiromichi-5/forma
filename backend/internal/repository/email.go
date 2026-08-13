package repository

import "context"

const (
	TemplateEmailVerification = "email-verification"
	TemplatePasswordReset     = "password-reset"
	TemplateInvite            = "invite"

	TemplateTicketStatusChanged         = "ticket-status-changed"
	TemplateTicketStatusChangedDetailed = "ticket-status-changed-detailed"
	TemplateTicketAssigned              = "ticket-assigned"
	TemplateTicketAssignedDetailed      = "ticket-assigned-detailed"
)

type SendEmailInput struct {
	To           []string
	TemplateName string
	TemplateData map[string]string
}

type EmailSender interface {
	SendEmail(ctx context.Context, input SendEmailInput) error
}
