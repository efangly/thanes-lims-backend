# Domain Glossary — Thanes LIMS Backend

## AI Chatbot (POC)

**Select AI** — Oracle Database 23ai feature (`DBMS_CLOUD_AI` package) that translates a natural-language question into SQL and executes it against tables in that same Oracle database, returning a narrated answer. Runs *inside* the Oracle DB, not as an external LLM API call from application code.

**AI Profile** — A named configuration stored in the Oracle DB (via `DBMS_CLOUD_AI.CREATE_PROFILE`) that tells Select AI which LLM provider to use (this project: **OCI Generative AI**), which tables/views are in scope, and what credentials to use for that provider.

**POC Oracle Instance** — A separate Oracle Autonomous Database (ADB), already provisioned, used *only* for this chatbot proof-of-concept. It is not the system of record — the SM-LIMS/Thanes LIMS backend's primary data store remains **Postgres** via GORM. The POC Oracle instance holds a synthetic, mirrored subset of the domain (Sample, TestResult, Inventory, PurchaseOrder) seeded independently — it is not synced from Postgres.

**Chatbot module** — New module in the existing hexagonal architecture (`internal/domain/chatbot`, etc.) exposing a single-turn `POST /chat`-style endpoint on the existing Go API, reusing the existing JWT auth. Calls into the POC Oracle instance via `godror` (Oracle wallet + Instant Client) to run Select AI queries and return the narrated answer.

## Storage Location

**Location** — A node in the physical storage hierarchy where samples are kept. Forms a tree: a Location may have a parent Location and child Locations. Every Location has a `Level Type` and a `Name`, and belongs to exactly one Cabinet's tree (its root ancestor).
_Avoid_: Place, position, spot (use Location for the entity; "leaf location" for the node a sample is actually stored at)

**Level Type** — The rung of the storage hierarchy a Location occupies: `cabinet` → `shelf` → `slot` → `sub_slot`, in that fixed order. A parent's Level Type must be the immediate predecessor of its child's — levels cannot be skipped.
_Avoid_: Level, type, tier

**Cabinet** — A root Location (`parent` is none, `level_type` is `cabinet`). Its Name is unique across the whole tree, not just among siblings.

**Leaf Location** — A Location with no children. Only a Leaf Location can be assigned to a Sample; any Level Type can be a leaf if the operator chooses not to subdivide it further (e.g. a Cabinet with no Shelves is itself a valid storage spot).

**Full Path** — The human-readable chain of a Location's ancestors down to itself (e.g. "Fridge-A / Shelf-2 / Slot-4"), derived on read from the tree — never stored.

## Access Control

**Role** — A named set of Permissions assigned to a User (exactly one Role per User): Admin, Lab Manager, QA, Scientist, or General. Determines what the User is allowed to do; carries no other meaning (not a job title or org-chart position).
_Avoid_: Group, level, tier

**Permission** — A grant of one Action on one Module to a Role (e.g. "Scientist may edit Sample"). Permissions are never assigned to a User directly, only via their Role.
_Avoid_: Right, scope, claim

**Action** — The verb half of a Permission: `view`, `create`, `edit`, `delete`, or `approve`. The same five Actions apply uniformly across every Module, even where one doesn't obviously make sense yet (e.g. `approve` on Document) — an unused Action is simply never granted to any Role, rather than the Module having a bespoke Action set. One exception: `export` exists only on the Audit Module, distinct from `view` — a Role can browse the Audit Trail (`view`) without being able to pull its PDF compliance export (`export`).

**Module** — The noun half of a Permission and of an Audit Entry's Resource: the business area being acted on (Sample, Location, Equipment, Inventory Item, Purchase Order, Document, Test Result, User, Environment, Notification, Audit). Corresponds to one hexagonal-architecture module in the codebase.

**Approve** — The Action that marks a record as reviewed and accepted by someone other than its author (e.g. QA approving a Test Result, or a Purchase Order). Distinct from `edit`: approving does not imply the ability to change the record's content.

## Audit Trail

**Audit Entry** — An automatically-recorded fact that a specific Actor performed a specific Action on a specific Module (and, for most Modules, a specific record) at a specific time. Written for every write, never edited or deleted after the fact.
_Avoid_: Audit log line, activity record

**Actor** — The authenticated User whose credentials were used to perform the Action an Audit Entry records. Every Audit Entry has exactly one Actor.

**Change Set** — The list of a record's fields whose values differed before and after a write, each as an old→new pair, attached to an Audit Entry. Only fields that actually changed appear — an Audit Entry for a write that touched one field carries a Change Set of one.
_Avoid_: Diff, delta

**Chain of Custody (CoC) Step** — A single handoff event in a Sample's custody history (who had it, when, what happened). Audited as its own Module, separate from the Sample it belongs to, because custody handoffs are themselves a compliance record independent of the Sample's other fields.

**Retired** — A record that has been deleted by a User but is kept in storage (never physically removed) so its Audit Trail and any references to it stay intact. Applies to every audited Module. A Retired record does not appear in normal listings and does not block reuse of values that were unique to it (e.g. a Retired User's email can be reused by a new User).
_Avoid_: Soft-deleted, archived

## Authentication & Sessions

**Session** — One continuous login on one device/browser, from initial Login until it is revoked or reaches its Absolute Session Lifetime. Represented in the system as a Token Family. A User may hold several Sessions at once (e.g. lab workstation and phone).
_Avoid_: Login, device (a Session is the login itself, not the hardware)

**Token Family** — The chain of Refresh Tokens produced by repeated Rolling Refresh of a single Session: each rotation revokes the presented token and issues its successor within the same family. Identified by a family id shared across the chain, with a `family_created_at` fixed at the first Login — used to enforce the Absolute Session Lifetime.
_Avoid_: Chain, token lineage

**Rolling Refresh** — The rule that using a valid Refresh Token immediately revokes it and issues a new one (same Token Family) with a fresh 7-day expiry — so an active Session never expires from inactivity alone, only by hitting the Absolute Session Lifetime or being revoked.

**Absolute Session Lifetime** — The hard cap of 30 days from a Token Family's `family_created_at`, after which Rolling Refresh stops renewing it regardless of activity and the User must Login again. Exists so no Session can persist forever purely by staying active.

**Reuse Detection** — The check that a Refresh Token already revoked by Rolling Refresh is being presented again — the signal that a Refresh Token has leaked. Triggers revocation of every Session the User holds, not just the offending Token Family, since a leaked token implies the whole account may be compromised.

## Caching

**Cache** — A Redis-backed, key-value read-through layer sitting in front of Postgres, accessed only through the `internal/ports/cache` port. Postgres remains the source of truth for every cached value; the Cache exists purely to reduce read load and is never the only place a value lives.
_Avoid_: Store, session store (Redis here is a Cache, not a data store of record)

**Refresh Token Cache Entry** — The cached copy of a Refresh Token row (family id, expiry, revoked flag, family_created_at), keyed by token hash, kept in sync with Postgres on every Rolling Refresh and revocation. Read fail-closed: if the Cache is unreachable, the request is rejected rather than risking a stale "not revoked" answer.
_Avoid_: Token blacklist (the cache holds live rows, not just a revoked-token denylist)

**Full Path Cache Entry** — The cached rendering of a Location's Full Path, keyed by Location id, expiring after a fixed TTL rather than being invalidated on Location changes. Read fail-open: if the Cache is unreachable or the entry has expired, Full Path is recomputed from Postgres directly.

**Refresh Cookie** — The httpOnly cookie that carries the current Refresh Token from browser to backend, scoped to the auth endpoints only. Distinct from the Access Token, which continues to travel as a bearer value in the response body / `Authorization` header, never as a cookie.
