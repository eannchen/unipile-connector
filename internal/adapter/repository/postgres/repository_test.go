package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRepositories(t *testing.T) {
	db := newTestDB(t)
	repos := GetRepositories(db)

	require.NotNil(t, repos.User)
	require.NotNil(t, repos.Account)
	require.NotNil(t, repos.Tx)

	require.IsType(t, (*accountRepo)(nil), repos.Account)
	require.IsType(t, (*userRepo)(nil), repos.User)
}
