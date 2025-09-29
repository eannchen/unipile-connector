package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
)

func TestUserRepository_CreateAndFetch(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entity.User{Username: "alice", Password: "hash"}
	require.NoError(t, repo.Create(ctx, user))

	fetchedByID, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", fetchedByID.Username)

	fetchedByUsername, err := repo.GetByUsername(ctx, "alice")
	require.NoError(t, err)
	require.Equal(t, user.ID, fetchedByUsername.ID)
}

func TestUserRepository_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	require.ErrorIs(t, err, repository.ErrRecordNotFound)

	_, err = repo.GetByUsername(ctx, "missing")
	require.ErrorIs(t, err, repository.ErrRecordNotFound)
}

func TestUserRepository_Create_Duplicate(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user1 := &entity.User{Username: "bob", Password: "hash"}
	require.NoError(t, repo.Create(ctx, user1))

	user2 := &entity.User{Username: "bob", Password: "hash2"}
	err := repo.Create(ctx, user2)
	require.Error(t, err)
}
