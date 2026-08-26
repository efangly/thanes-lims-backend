package cachedenvironment_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedenvironment"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) List(ctx context.Context) ([]environment.Gauge, error) {
	args := m.Called(ctx)
	var v []environment.Gauge
	if g := args.Get(0); g != nil {
		v = g.([]environment.Gauge)
	}
	return v, args.Error(1)
}
func (m *mockRepo) FindByLocation(ctx context.Context, location string) (environment.Gauge, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(environment.Gauge), args.Error(1)
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

func encode(t *testing.T, g environment.Gauge) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(g))
	return buf.Bytes()
}

func TestFindByLocation_CacheHit_SkipsRepository(t *testing.T) {
	g := environment.Gauge{Location: "Fridge-A", Unit: "C", RangeMin: 2, RangeMax: 8}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "env:gauge:Fridge-A").Return(encode(t, g), nil)

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	got, err := sut.FindByLocation(context.Background(), "Fridge-A")

	require.NoError(t, err)
	assert.Equal(t, g, got)
	repo.AssertNotCalled(t, "FindByLocation", mock.Anything, mock.Anything)
}

func TestFindByLocation_CacheMiss_FallsBackAndPopulates(t *testing.T) {
	g := environment.Gauge{Location: "Freezer-1", Unit: "C", RangeMin: -20, RangeMax: -15}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "env:gauge:Freezer-1").Return(nil, shared.ErrNotFound)
	repo.On("FindByLocation", mock.Anything, "Freezer-1").Return(g, nil)
	c.On("Set", mock.Anything, "env:gauge:Freezer-1", encode(t, g), 15*time.Minute).Return(nil)

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	got, err := sut.FindByLocation(context.Background(), "Freezer-1")

	require.NoError(t, err)
	assert.Equal(t, g, got)
	c.AssertExpectations(t)
}

// Unlike the Refresh Token cache, this cache is fail-open: a genuine cache
// error (not just a miss) still falls back to Postgres rather than
// rejecting the request (see ADR 0006).
func TestFindByLocation_CacheUnreachable_FallsBackToRepository(t *testing.T) {
	g := environment.Gauge{Location: "Incubator-2", Unit: "C", RangeMin: 35, RangeMax: 39}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "env:gauge:Incubator-2").Return(nil, errors.New("connection refused"))
	repo.On("FindByLocation", mock.Anything, "Incubator-2").Return(g, nil)
	c.On("Set", mock.Anything, "env:gauge:Incubator-2", encode(t, g), 15*time.Minute).Return(nil)

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	got, err := sut.FindByLocation(context.Background(), "Incubator-2")

	require.NoError(t, err)
	assert.Equal(t, g, got)
}

func TestFindByLocation_RepositoryNotFound_Propagates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "env:gauge:Unknown").Return(nil, shared.ErrNotFound)
	repo.On("FindByLocation", mock.Anything, "Unknown").Return(environment.Gauge{}, shared.ErrNotFound)

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	_, err := sut.FindByLocation(context.Background(), "Unknown")

	assert.ErrorIs(t, err, shared.ErrNotFound)
	c.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFindByLocation_CacheSetFailure_StillReturnsResult(t *testing.T) {
	g := environment.Gauge{Location: "Fridge-B", Unit: "C", RangeMin: 2, RangeMax: 8}
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "env:gauge:Fridge-B").Return(nil, shared.ErrNotFound)
	repo.On("FindByLocation", mock.Anything, "Fridge-B").Return(g, nil)
	c.On("Set", mock.Anything, "env:gauge:Fridge-B", encode(t, g), 15*time.Minute).Return(errors.New("connection refused"))

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	got, err := sut.FindByLocation(context.Background(), "Fridge-B")

	require.NoError(t, err)
	assert.Equal(t, g, got)
}

func TestList_PassesThrough(t *testing.T) {
	gauges := []environment.Gauge{{Location: "Fridge-A", Unit: "C", RangeMin: 2, RangeMax: 8}}
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("List", mock.Anything).Return(gauges, nil)

	sut := cachedenvironment.NewCachedGaugeRepository(repo, c)
	got, err := sut.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, gauges, got)
	c.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}
