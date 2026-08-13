//go:build integration

package user_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := user.New(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domainuser.User{
		Name:         "Somchai Test",
		Email:        "somchai@example.com",
		PasswordHash: "hashed-password",
		Role:         domainuser.RoleScientist,
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, "somchai@example.com", created.Email)

	byID, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byID.ID)
	assert.Equal(t, created.Name, byID.Name)
	assert.Equal(t, created.Email, byID.Email)
	assert.Equal(t, created.Role, byID.Role)

	byEmail, err := repo.FindByEmail(ctx, "somchai@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byEmail.ID)

	all, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := created
	updated.Name = "Somchai Updated"
	updated.Role = domainuser.RoleAdmin
	result, err := repo.Update(ctx, updated)
	require.NoError(t, err)
	assert.Equal(t, "Somchai Updated", result.Name)
	assert.Equal(t, domainuser.RoleAdmin, result.Role)
}

func TestUserRepository_NotFound(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := user.New(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999999)
	assert.ErrorIs(t, err, shared.ErrNotFound)

	_, err = repo.FindByEmail(ctx, "does-not-exist@example.com")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}
