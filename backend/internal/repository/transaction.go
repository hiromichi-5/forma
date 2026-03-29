package repository

import "context"

type Transaction interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
