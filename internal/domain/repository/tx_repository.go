package repository

import "context"

// TxRepository is a repository that can execute transactions
type TxRepository interface {
	Do(ctx context.Context, fn func(repos *Repositories) error) error
}
