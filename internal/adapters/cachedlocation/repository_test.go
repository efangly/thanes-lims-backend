package cachedlocation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedlocation"
	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, l location.Location) (location.Location, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockRepo) CreateMany(ctx context.Context, ls []location.Location) ([]location.Location, error) {
	args := m.Called(ctx, ls)
	return args.Get(0).([]location.Location), args.Error(1)
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (location.Location, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockRepo) ListChildren(ctx context.Context, parentID *string) ([]location.Location, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]location.Location), args.Error(1)
}
func (m *mockRepo) FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error) {
	args := m.Called(ctx, parentID, name)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockRepo) HasChildren(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRepo) FullPath(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
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

func TestFullPath_CacheHit_SkipsRepository(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "location:fullpath:LOC-1").Return([]byte("Fridge-A / Shelf-2"), nil)

	sut := cachedlocation.NewCachedRepository(repo, c)
	got, err := sut.FullPath(context.Background(), "LOC-1")

	require.NoError(t, err)
	assert.Equal(t, "Fridge-A / Shelf-2", got)
	repo.AssertNotCalled(t, "FullPath", mock.Anything, mock.Anything)
}

func TestFullPath_CacheMiss_FallsBackAndPopulates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "location:fullpath:LOC-2").Return(nil, shared.ErrNotFound)
	repo.On("FullPath", mock.Anything, "LOC-2").Return("Fridge-A / Shelf-3", nil)
	c.On("Set", mock.Anything, "location:fullpath:LOC-2", []byte("Fridge-A / Shelf-3"), 15*time.Minute).Return(nil)

	sut := cachedlocation.NewCachedRepository(repo, c)
	got, err := sut.FullPath(context.Background(), "LOC-2")

	require.NoError(t, err)
	assert.Equal(t, "Fridge-A / Shelf-3", got)
	c.AssertExpectations(t)
}

// Unlike the Refresh Token cache, this cache is fail-open: a genuine cache
// error (not just a miss) still falls back to Postgres rather than
// rejecting the request (see ADR 0005).
func TestFullPath_CacheUnreachable_FallsBackToRepository(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "location:fullpath:LOC-3").Return(nil, errors.New("connection refused"))
	repo.On("FullPath", mock.Anything, "LOC-3").Return("Fridge-A", nil)
	c.On("Set", mock.Anything, "location:fullpath:LOC-3", []byte("Fridge-A"), 15*time.Minute).Return(nil)

	sut := cachedlocation.NewCachedRepository(repo, c)
	got, err := sut.FullPath(context.Background(), "LOC-3")

	require.NoError(t, err)
	assert.Equal(t, "Fridge-A", got)
}

func TestFullPath_RepositoryNotFound_Propagates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "location:fullpath:LOC-4").Return(nil, shared.ErrNotFound)
	repo.On("FullPath", mock.Anything, "LOC-4").Return("", shared.ErrNotFound)

	sut := cachedlocation.NewCachedRepository(repo, c)
	_, err := sut.FullPath(context.Background(), "LOC-4")

	assert.ErrorIs(t, err, shared.ErrNotFound)
	c.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFullPath_CacheSetFailure_StillReturnsResult(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "location:fullpath:LOC-5").Return(nil, shared.ErrNotFound)
	repo.On("FullPath", mock.Anything, "LOC-5").Return("Fridge-B", nil)
	c.On("Set", mock.Anything, "location:fullpath:LOC-5", []byte("Fridge-B"), 15*time.Minute).Return(errors.New("connection refused"))

	sut := cachedlocation.NewCachedRepository(repo, c)
	got, err := sut.FullPath(context.Background(), "LOC-5")

	require.NoError(t, err)
	assert.Equal(t, "Fridge-B", got)
}
