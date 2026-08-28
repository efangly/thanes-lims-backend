package cacheduser_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cacheduser"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, rt user.RefreshToken) (user.RefreshToken, error) {
	args := m.Called(ctx, rt)
	return args.Get(0).(user.RefreshToken), args.Error(1)
}
func (m *mockRepo) FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(user.RefreshToken), args.Error(1)
}
func (m *mockRepo) Revoke(ctx context.Context, id int64, tokenHash string) (int64, error) {
	args := m.Called(ctx, id, tokenHash)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockRepo) RevokeAllForUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *mockRepo) FindTokenHashesByUserID(ctx context.Context, userID int64) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
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

func TestFindByTokenHash_CacheHit_SkipsRepository(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	stored := user.RefreshToken{ID: 1, TokenHash: "hash1", ExpiresAt: time.Now().Add(time.Hour)}
	data, _ := json.Marshal(stored)
	c.On("Get", mock.Anything, "refresh:hash1").Return(data, nil)

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	got, err := sut.FindByTokenHash(context.Background(), "hash1")

	require.NoError(t, err)
	assert.Equal(t, stored.ID, got.ID)
	repo.AssertNotCalled(t, "FindByTokenHash", mock.Anything, mock.Anything)
}

func TestFindByTokenHash_CacheMiss_FallsBackAndPopulates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	stored := user.RefreshToken{ID: 2, TokenHash: "hash2", ExpiresAt: time.Now().Add(time.Hour)}
	c.On("Get", mock.Anything, "refresh:hash2").Return(nil, shared.ErrNotFound)
	repo.On("FindByTokenHash", mock.Anything, "hash2").Return(stored, nil)
	c.On("Set", mock.Anything, "refresh:hash2", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	got, err := sut.FindByTokenHash(context.Background(), "hash2")

	require.NoError(t, err)
	assert.Equal(t, stored.ID, got.ID)
	c.AssertExpectations(t)
}

func TestFindByTokenHash_CacheUnreachable_FailsClosed(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	c.On("Get", mock.Anything, "refresh:hash3").Return(nil, errors.New("connection refused"))

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	_, err := sut.FindByTokenHash(context.Background(), "hash3")

	require.Error(t, err)
	repo.AssertNotCalled(t, "FindByTokenHash", mock.Anything, mock.Anything)
}

func TestRevoke_DeletesCacheAfterPostgres(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("Revoke", mock.Anything, int64(5), "hash5").Return(int64(1), nil)
	c.On("Delete", mock.Anything, "refresh:hash5").Return(nil)

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	_, err := sut.Revoke(context.Background(), 5, "hash5")

	require.NoError(t, err)
	c.AssertExpectations(t)
}

func TestRevoke_CacheDeleteFailure_Propagates(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("Revoke", mock.Anything, int64(6), "hash6").Return(int64(1), nil)
	c.On("Delete", mock.Anything, "refresh:hash6").Return(errors.New("connection refused"))

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	_, err := sut.Revoke(context.Background(), 6, "hash6")

	assert.Error(t, err)
}

func TestRevoke_PostgresFailure_NeverTouchesCache(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("Revoke", mock.Anything, int64(7), "hash7").Return(int64(0), errors.New("db down"))

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	_, err := sut.Revoke(context.Background(), 7, "hash7")

	assert.Error(t, err)
	c.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestRevokeAllForUser_InvalidatesEveryHash(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("FindTokenHashesByUserID", mock.Anything, int64(9)).Return([]string{"h1", "h2"}, nil)
	repo.On("RevokeAllForUser", mock.Anything, int64(9)).Return(nil)
	c.On("Delete", mock.Anything, "refresh:h1").Return(nil)
	c.On("Delete", mock.Anything, "refresh:h2").Return(nil)

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	err := sut.RevokeAllForUser(context.Background(), 9)

	require.NoError(t, err)
	c.AssertExpectations(t)
}

func TestRevokeAllForUser_CacheDeleteFailure_AttemptsAllThenErrors(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	repo.On("FindTokenHashesByUserID", mock.Anything, int64(9)).Return([]string{"h1", "h2"}, nil)
	repo.On("RevokeAllForUser", mock.Anything, int64(9)).Return(nil)
	// h1 fails, but h2 must still be attempted; the call then reports an error
	// so the security path (Reuse Detection / global logout) can alert.
	c.On("Delete", mock.Anything, "refresh:h1").Return(errors.New("connection refused"))
	c.On("Delete", mock.Anything, "refresh:h2").Return(nil)

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	err := sut.RevokeAllForUser(context.Background(), 9)

	require.Error(t, err)
	c.AssertExpectations(t)
}

func TestCreate_PopulatesCache_IgnoresSetFailure(t *testing.T) {
	c := new(mockCache)
	repo := new(mockRepo)
	created := user.RefreshToken{ID: 3, TokenHash: "hash4", ExpiresAt: time.Now().Add(time.Hour)}
	repo.On("Create", mock.Anything, mock.AnythingOfType("user.RefreshToken")).Return(created, nil)
	c.On("Set", mock.Anything, "refresh:hash4", mock.Anything, mock.AnythingOfType("time.Duration")).Return(errors.New("connection refused"))

	sut := cacheduser.NewCachedRefreshTokenRepository(repo, c)
	got, err := sut.Create(context.Background(), user.RefreshToken{})

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}
