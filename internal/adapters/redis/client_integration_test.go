//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/redis/rtest"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_GetSetDelete(t *testing.T) {
	c := rtest.SetupRedis(t)
	ctx := context.Background()

	_, err := c.Get(ctx, "missing")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	require.NoError(t, c.Set(ctx, "key1", []byte("value1"), time.Minute))
	got, err := c.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), got)

	require.NoError(t, c.Delete(ctx, "key1"))
	_, err = c.Get(ctx, "key1")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestAdapter_SetWithTTL_ExpiresEntry(t *testing.T) {
	c := rtest.SetupRedis(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "short-lived", []byte("v"), 50*time.Millisecond))
	time.Sleep(200 * time.Millisecond)

	_, err := c.Get(ctx, "short-lived")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}
