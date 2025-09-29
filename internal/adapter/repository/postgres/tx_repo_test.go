package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/repository"
)

func TestTxRepository_Do(t *testing.T) {
	db := newTestDB(t)
	repos := repository.Repositories{}
	txRepo := NewTxRepository(db, &repos)

	ctx := context.Background()

	var called bool
	err := txRepo.Do(ctx, func(r *repository.Repositories) error {
		called = true
		if r != &repos {
			t.Fatalf("unexpected repositories pointer")
		}
		return nil
	})

	require.NoError(t, err)
	require.True(t, called)
}
