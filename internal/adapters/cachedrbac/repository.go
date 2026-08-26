// Package cachedrbac decorates a ports/rbac.Repository, adding a Redis
// read-through Cache in front of FindPermissionsByRoleName, per
// docs/adr/0006-redis-cache-for-rbac-permissions-and-gauge-thresholds.md.
package cachedrbac

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	portrbac "github.com/efangly/thanes-lims-backend/internal/ports/rbac"
)

// permsTTL matches JWT_ACCESS_TTL: Role/Permission assignment is
// seed/migration-only in this phase (no endpoint edits role_permissions),
// so a short, fixed staleness window needs no invalidation path. This
// cache is fail-open: any cache miss or error, including Redis being
// unreachable, simply falls back to Postgres exactly as before (see ADR
// 0006).
const permsTTL = 15 * time.Minute

type CachedRepository struct {
	next  portrbac.Repository
	cache cache.Cache
}

func NewCachedRepository(next portrbac.Repository, c cache.Cache) *CachedRepository {
	return &CachedRepository{next: next, cache: c}
}

func permsCacheKey(roleName string) string {
	return "rbac:perms:" + roleName
}

// FindPermissionsByRoleName returns the cached permission set when present;
// any cache miss or error falls straight back to Postgres, matching
// pre-cache behavior.
func (r *CachedRepository) FindPermissionsByRoleName(ctx context.Context, roleName string) ([]rbac.Permission, error) {
	key := permsCacheKey(roleName)

	if data, err := r.cache.Get(ctx, key); err == nil {
		var perms []rbac.Permission
		if decErr := gob.NewDecoder(bytes.NewReader(data)).Decode(&perms); decErr == nil {
			return perms, nil
		}
	}

	perms, err := r.next.FindPermissionsByRoleName(ctx, roleName)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if encErr := gob.NewEncoder(&buf).Encode(perms); encErr != nil {
		log.Printf("cachedrbac: encode permissions for cache: %v", encErr)
		return perms, nil
	}
	if err := r.cache.Set(ctx, key, buf.Bytes(), permsTTL); err != nil {
		log.Printf("cachedrbac: populate permissions cache: %v", err)
	}
	return perms, nil
}
