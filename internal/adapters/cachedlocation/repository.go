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
//
// TODO(location-invalidation): Delete already evicts its own key; a
// rename/move only takes effect after fullPathTTL and also leaves stale
// entries for every descendant. If that window becomes unacceptable, evict
// the subtree (or version-prefix the keys) on rename/move.
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

func (r *CachedRepository) ListRoots(ctx context.Context, kind location.Kind) ([]location.Location, error) {
	return r.next.ListRoots(ctx, kind)
}

func (r *CachedRepository) FindByBarcode(ctx context.Context, code string) (location.Location, error) {
	return r.next.FindByBarcode(ctx, code)
}

func (r *CachedRepository) FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error) {
	return r.next.FindChildByName(ctx, parentID, name)
}

func (r *CachedRepository) HasChildren(ctx context.Context, id string) (bool, error) {
	return r.next.HasChildren(ctx, id)
}

// Delete removes the Location, then best-effort evicts its Full Path cache
// entry so a deleted Location's path stops being served immediately rather
// than lingering for up to fullPathTTL. A failed eviction is only logged:
// this cache is fail-open and the entry expires on its own TTL regardless.
// (Renames/moves are still only bounded by fullPathTTL - see ADR 0005.)
func (r *CachedRepository) Delete(ctx context.Context, id string) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	if err := r.cache.Delete(ctx, fullPathCacheKey(id)); err != nil {
		log.Printf("cachedlocation: evict full path cache on delete: %v", err)
	}
	return nil
}
