package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
)

func TestAccountRepository_CreateAndRetrieve(t *testing.T) {
	db := newTestDB(t)
	repo := NewAccountRepository(db)

	account := &entity.Account{
		UserID:        1,
		Provider:      "LINKEDIN",
		AccountID:     "acc-123",
		CurrentStatus: "OK",
	}
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, account))

	accounts, err := repo.GetByUserID(ctx, 1)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.Equal(t, "acc-123", accounts[0].AccountID)

	got, err := repo.GetByUserIDAndAccountIDForUpdate(ctx, 1, "acc-123")
	require.NoError(t, err)
	require.Equal(t, account.AccountID, got.AccountID)
}

func TestAccountRepository_GetByUserIDAndAccountID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewAccountRepository(db)

	_, err := repo.GetByUserIDAndAccountIDForUpdate(context.Background(), 99, "missing")
	require.ErrorIs(t, err, repository.ErrAccountNotFound)
}

func TestAccountRepository_UpdateAndDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()

	account := &entity.Account{
		UserID:        2,
		Provider:      "LINKEDIN",
		AccountID:     "acc-456",
		CurrentStatus: "PENDING",
	}
	require.NoError(t, repo.Create(ctx, account))

	account.CurrentStatus = "OK"
	require.NoError(t, repo.Update(ctx, account))

	updated, err := repo.GetByUserIDAndAccountIDForUpdate(ctx, 2, "acc-456")
	require.NoError(t, err)
	require.Equal(t, "OK", updated.CurrentStatus)

	require.NoError(t, repo.DeleteByUserIDAndAccountID(ctx, 2, "acc-456"))

	accounts, err := repo.GetByUserID(ctx, 2)
	require.NoError(t, err)
	require.Len(t, accounts, 0)
}
