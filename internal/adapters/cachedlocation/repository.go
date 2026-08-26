// Package cachedlocation decorates a ports/location.LocationRepository,
// adding a Redis read-through Cache in front of FullPath only, per
// docs/adr/0005-redis-cache-for-refresh-tokens-and-location-full-path.md.
// Every other method passes straight through - Full Path is the only
// expensive, hot-path recursive query in this module.
package cachedlocation

import (
	"context"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

// fullPathTTL matches JWT_ACCESS_TTL: Locations change rarely once a
// Cabinet's tree is set up, so a short, fixed staleness window is an
// acceptable trade against invalidating on every rename/move/delete (see
// ADR 0005). This cache is fail-open: any cache miss or error, including
// Redis being unreachable, simply falls back to Postgres exactly as before.
const fullPathTTL = 15 * time.Minute

type CachedRepository struct {
	next  portlocation.LocationRepository
	cache cache.Cache
}

func NewCachedRepository(next portlocation.LocationRepository, c cache.Cache) *CachedRepository {
	return &CachedRepository{next: next, cache: c}
}

func fullPathCacheKey(id string) string {
	return "location:fullpath:" + id
}

// FullPath returns the cached rendering when present; any cache miss or
// error falls straight back to Postgres, matching pre-cache behavior.
func (r *CachedRepository) FullPath(ctx context.Context, id string) (string, error) {
	key := fullPathCacheKey(id)

	if data, err := r.cache.Get(ctx, key); err == nil {
		return string(data), nil
	}

	fullPath, err := r.next.FullPath(ctx, id)
	if err != nil {
		return "", err
	}
	if err := r.cache.Set(ctx, key, []byte(fullPath), fullPathTTL); err != nil {
		log.Printf("cachedlocation: populate full path cache: %v", err)
	}
	return fullPath, nil
}

func (r *CachedRepository) Create(ctx context.Context, l location.Location) (location.Location, error) {
	return r.next.Create(ctx, l)
}

func (r *CachedRepository) CreateMany(ctx context.Context, ls []location.Location) ([]location.Location, error) {
	return r.next.CreateMany(ctx, ls)
}

func (r *CachedRepository) GetByID(ctx context.Context, id string) (location.Location, error) {
	return r.next.GetByID(ctx, id)
}

func (r *CachedRepository) ListChildren(ctx context.Context, parentID *string) ([]location.Location, error) {
	return r.next.ListChildren(ctx, parentID)
}

func (r *CachedRepository) FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error) {
	return r.next.FindChildByName(ctx, parentID, name)
}

func (r *CachedRepository) HasChildren(ctx context.Context, id string) (bool, error) {
	return r.next.HasChildren(ctx, id)
}

func (r *CachedRepository) Delete(ctx context.Context, id string) error {
	return r.next.Delete(ctx, id)
}
