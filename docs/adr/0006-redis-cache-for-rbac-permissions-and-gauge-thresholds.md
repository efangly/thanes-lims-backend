# Redis Cache for RBAC Permissions and Gauge Thresholds

[ADR 0005](./0005-redis-cache-for-refresh-tokens-and-location-full-path.md) wired Redis in as a read-through Cache behind the generic `internal/ports/cache.Cache` port for two hot paths: Refresh Token lookups and Location Full Path. A follow-up survey of the rest of the codebase (every domain outside auth: sample, rbac, inventory, equipment, purchaseorder, testresult, document, notification, audit, environment, location, user) found two more reads that fit the same shape — hot path, non-trivial query, rarely-changing underlying data — and no others worth it yet.

## Decision

Extend the same Cache port to two more read paths, both fail-open and TTL-only like Location Full Path (not fail-closed like the Refresh Token entry) — see rationale per entry below.

### RBAC Permission Cache Entry

`rbac.Repository.FindPermissionsByRoleName` joins `permissions` × `role_permissions` × `roles` by role name. It is called on every Login and every `/auth/refresh` — the exact same request path already carrying the Refresh Token cache lookup from ADR 0005 — so this closes a gap left by that ADR rather than opening new scope.

- Key: `rbac:perms:{role_name}`
- TTL: 15 minutes, matching `JWT_ACCESS_TTL`
- Fail-open: a cache miss or Redis outage falls back to Postgres, same as before this change existed
- No invalidation: there is no endpoint in the system that edits `role_permissions` — Role/Permission assignment is seed/migration-only in this phase (per [ADR 0002](./0002-normalized-rbac-with-jwt-embedded-permissions.md)), so TTL alone is sufficient. If a role-editing endpoint is added later, it must `Delete` the affected `rbac:perms:{role_name}` key(s) or this cache will silently serve stale permissions.

### Gauge Threshold Cache Entry

`environment.GaugeRepository.FindByLocation` is called from `EvaluateThresholdsUseCase`, itself invoked from `RecordReadingUseCase` on every sensor reading ingested via `POST /environment/readings`. Gauge threshold rows (`RangeMin`/`RangeMax`, `Unit`) are configuration set up per Location and changed only by recalibration.

Actual sensor reporting frequency is not specified anywhere in this codebase or its docs (no polling-interval env var, no batching — the endpoint takes one reading per call). This entry is included on the shape of the call path (once per ingested reading, unbounded by us) rather than on a measured rate, mirroring the judgment ADR 0005 already made for Location Full Path ("neither is expensive enough on its own to force this change... we're wiring it in now rather than waiting for either to become a measured problem").

- Key: `env:gauge:{location}`
- TTL: 15 minutes, same window as the other two entries
- Fail-open: a cache miss or Redis outage falls back to Postgres
- No invalidation: there is no endpoint that creates or updates a `gauges` row either (only `List`), so the same TTL-only reasoning applies

### Ruled out

- **`notification.Repository.ListForUser`** — underlying data changes on every `Create`/`MarkRead`/`MarkAllRead`, the opposite of "rarely changes." Caching would add invalidation complexity for a single indexed-column query that's already cheap.
- **`environment.ReadingRepository.LatestByLocation` / `ListTrend`** — sensor readings are continuously-changing data by design; staleness tolerance is inverted versus Gauge thresholds.
- **`user.Repository.roleIDByRole`** — only called on User Create/Update (admin, low frequency), not a hot path.
- Every other repository surveyed (inventory, equipment, purchaseorder, testresult, document, sample CoC steps) is plain single-table or single-join CRUD with indexed lookups — no aggregation-on-read, no recursive query, no N+1 pattern across a list response.

### Watch item, not in scope

Audit Trail entries currently expose only `ActorID`, no hydrated actor display name — so there's no N+1 user-lookup to cache today. If a future change adds actor-name hydration to `audit.Repository.List` or its DTOs, that would become a `user:name:{id}`-style caching candidate at that point, not before.

## Consequences

Same operational shape as ADR 0005: one Redis instance, one generic port, one adapter package per decorated repository (`internal/adapters/cachedrbac`, `internal/adapters/cachedenvironment`), each wrapping only the one method that benefits and passing every other method straight through. No new client library, no new config, no `docker-compose.yml` change — reuses the existing `REDIS_URL` connection already provisioned for ADR 0005.

**Status**: accepted
