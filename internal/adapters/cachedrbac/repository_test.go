package cachedrbac_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedrbac"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) FindPermissionsByRoleName(ctx context.Context, roleName string) ([]rbac.Permission, error) {
	args := m.Called(ctx, roleName)
	var perms []rbac.Permission
	if v := args.Get(0); v != nil {
		perms = v.([]rbac.Permission)
	}
	return perms, args.Error(1)
}

type mockCache struct{ mock.Mock }

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	var v []byte
	if b := args.Get(0); b != nil {
		v = b.([]byte)
	}
	return v, args.Error(1)
}
func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func encode(t *testing.T, perms []rbac.Permission) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(perms))
	return buf.Bytes()
}

func TestFindPermissionsByRoleName_CacheHit_SkipsRepository(t *testing.T) {
	perms := []rbac.Permission{{ID: 1, Module: rbac.Module("sample"), Action: rbac.Action("view")}}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "rbac:perms:Scientist").Return(encode(t, perms), nil)

	sut := cachedrbac.NewCachedRepository(repo, c)
	got, err := sut.FindPermissionsByRoleName(context.Background(), "Scientist")

	require.NoError(t, err)
	assert.Equal(t, perms, got)
	repo.AssertNotCalled(t, "FindPermissionsByRoleName", mock.Anything, mock.Anything)
}

func TestFindPermissionsByRoleName_CacheMiss_FallsBackAndPopulates(t *testing.T) {
	perms := []rbac.Permission{{ID: 2, Module: rbac.Module("location"), Action: rbac.Action("edit")}}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "rbac:perms:Admin").Return(nil, shared.ErrNotFound)
	repo.On("FindPermissionsByRoleName", mock.Anything, "Admin").Return(perms, nil)
	c.On("Set", mock.Anything, "rbac:perms:Admin", encode(t, perms), 15*time.Minute).Return(nil)

	sut := cachedrbac.NewCachedRepository(repo, c)
	got, err := sut.FindPermissionsByRoleName(context.Background(), "Admin")

	require.NoError(t, err)
	assert.Equal(t, perms, got)
	c.AssertExpectations(t)
}

// Unlike the Refresh Token cache, this cache is fail-open: a genuine cache
// error (not just a miss) still falls back to Postgres rather than
// rejecting the request (see ADR 0006).
func TestFindPermissionsByRoleName_CacheUnreachable_FallsBackToRepository(t *testing.T) {
	perms := []rbac.Permission{{ID: 3, Module: rbac.Module("audit"), Action: rbac.Action("view")}}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "rbac:perms:QA").Return(nil, errors.New("connection refused"))
	repo.On("FindPermissionsByRoleName", mock.Anything, "QA").Return(perms, nil)
	c.On("Set", mock.Anything, "rbac:perms:QA", encode(t, perms), 15*time.Minute).Return(nil)

	sut := cachedrbac.NewCachedRepository(repo, c)
	got, err := sut.FindPermissionsByRoleName(context.Background(), "QA")

	require.NoError(t, err)
	assert.Equal(t, perms, got)
}

func TestFindPermissionsByRoleName_RepositoryError_Propagates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "rbac:perms:Unknown").Return(nil, shared.ErrNotFound)
	repo.On("FindPermissionsByRoleName", mock.Anything, "Unknown").Return(nil, errors.New("boom"))

	sut := cachedrbac.NewCachedRepository(repo, c)
	_, err := sut.FindPermissionsByRoleName(context.Background(), "Unknown")

	assert.Error(t, err)
	c.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFindPermissionsByRoleName_CacheSetFailure_StillReturnsResult(t *testing.T) {
	perms := []rbac.Permission{{ID: 4, Module: rbac.Module("equipment"), Action: rbac.Action("create")}}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "rbac:perms:General").Return(nil, shared.ErrNotFound)
	repo.On("FindPermissionsByRoleName", mock.Anything, "General").Return(perms, nil)
	c.On("Set", mock.Anything, "rbac:perms:General", encode(t, perms), 15*time.Minute).Return(errors.New("connection refused"))

	sut := cachedrbac.NewCachedRepository(repo, c)
	got, err := sut.FindPermissionsByRoleName(context.Background(), "General")

	require.NoError(t, err)
	assert.Equal(t, perms, got)
}
