package repository

import "context"

type TxManager interface {
	Do(ctx context.Context, fn func(repos Repos) error) error
}

type Repos struct {
	User   UserRepository
	Form   FormRepository
	Member MemberRepository
	Status StatusRepository
	Invite InviteRepository
	Ticket TicketRepository
}
