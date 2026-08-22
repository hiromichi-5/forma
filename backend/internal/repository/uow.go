package repository

import "context"

type UnitOfWork[T any] interface {
	Do(ctx context.Context, fn func(repos T) error) error
}

type AuthRepos struct {
	User                   UserRepository
	EmailVerificationToken EmailVerificationTokenRepository
	PasswordResetToken     PasswordResetTokenRepository
}

type FormRepos struct {
	Form   FormRepository
	Member MemberRepository
	Status StatusRepository
}

type InviteRepos struct {
	Invite InviteRepository
	User   UserRepository
	Member MemberRepository
}

type StatusRepos struct {
	Status StatusRepository
}

type TicketRepos struct {
	Ticket TicketRepository
	Status StatusRepository
	User   UserRepository
}

type NotificationRepos struct {
	Notification NotificationRepository
}
