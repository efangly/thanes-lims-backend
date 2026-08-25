# Normalized RBAC with JWT-embedded permissions, not a code-level role matrix

Authorization today is `internal/domain/user/role.go`: a fixed `Role` enum (`admin`/`qa`/`scientist`/`general`) with a hard-coded, global `Can(perm)` matrix of four verbs (`view`/`edit`/`delete`/`approve`) — not scoped per Module, and only actually enforced on `/users` and `/audit/export`; every other Module route checks authentication but not authorization. We're replacing this with `roles`, `permissions`, and `role_permissions` tables (`users.role_id` FK), where a Permission is a `module:action` pair, and enforcing it on every Module route.

We picked DB-normalized Roles/Permissions over extending the Go matrix because Permissions now need to be independently grantable per Module (e.g. Scientist can edit Sample but not Location) — a shape the flat four-verb matrix can't express without becoming a large switch statement, and because adding a Module or a Role should not require a code change and redeploy. The cost: a Permission check now depends on data, not just code, so we embed the Actor's resolved Permission set into the JWT access token at login rather than querying `role_permissions` on every request. This trades a bounded staleness window (permission changes take effect within one access-token TTL, currently 15 minutes) for avoiding a DB round-trip per request; a Role or Permission edit revokes the affected User's refresh tokens so that window is a hard ceiling, not best-effort.

## Status

accepted

## Considered Options

- **Keep the Go-code role matrix, add Module scoping to it** — no schema change, but every new Module or Role still requires a code change and redeploy, and the matrix was already unscoped/unenforced almost everywhere; doesn't fix the root problem.
- **Normalized tables + per-request DB lookup for permission checks** — always fresh, no staleness window, but adds a DB call (or a cache with its own invalidation problem) to every authenticated request.
- **Normalized tables + JWT-embedded permissions (chosen)** — no per-request DB cost; staleness is bounded by access-token TTL and closed early by revoking refresh tokens on change.
