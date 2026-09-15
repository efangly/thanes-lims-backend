# Partner API: gRPC transport, and adding ListDevicesByWard

SMtrack retired the REST version of its Partner API (docs/partner-api-guide.md, previous
revision) and replaced it with a gRPC service, `partner.PartnerService`
(`proto/partner/partner.proto`). This backend's outbound integration (`internal/adapters/partnergrpc`,
formerly `internal/adapters/partnerapi`) has been rewritten against the new contract — there
is no REST fallback to keep, since SMtrack no longer exposes one.

## What changed

- Auth moves from a static `X-API-Key` HTTP header to a per-call `x-api-key` gRPC metadata
  value (`metadata.AppendToOutgoingContext`).
- Errors move from HTTP status codes wrapped in a `{success,message,data}` envelope to
  standard gRPC status codes (`codes.Code`). The retryable/permanent classification this
  backend already made for stale-cache fallback (ADR 0011) is preserved, just remapped:
  `UNAUTHENTICATED`/`NOT_FOUND`/`INVALID_ARGUMENT` → permanent;
  `RESOURCE_EXHAUSTED`/`UNAVAILABLE`/`DEADLINE_EXCEEDED` → retryable.
- `GetDeviceSnapshot` now returns metadata **and** the trailing 1-hour telemetry window in a
  single call, replacing the old two-REST-call pattern (a metadata call, then a
  timeseries call for the latest reading). The `PartnerAPIClient` port collapses to one
  `FetchSnapshot` method accordingly.
- Connection is plaintext (`insecure.NewCredentials()`) — SMtrack has not yet configured TLS
  for this endpoint. This is a known, accepted gap, not a deliberate security choice; revisit
  when SMtrack adds TLS.

## New capability: ListDevicesByWard

The new proto also exposes `ListDevicesByWard(ward, page, limit)`, with no REST precedent —
this backend's old integration only ever fetched by a single, already-known `Serial` (device
discovery/mapping was entirely manual, via the `PartnerDevice` CRUD endpoints). We added a
new admin-facing endpoint, `GET /partner-devices/discover?ward=...`, backed by this RPC, so
an admin can browse SMtrack's device inventory in a ward to find a device's `serial` before
creating the Serial↔Location mapping — rather than having to get it from SMtrack out of
band. Like every other Partner Device read, this is read-through only; nothing from this RPC
is persisted (see ADR 0011).

## Not superseded: ADR 0011

ADR 0011's core decision — proxy Partner Device readings live, cache them briefly, never
persist raw telemetry to `sensor_readings` — is transport-agnostic and unchanged by this
switch. This ADR only changes *how* the proxying talks to SMtrack.

One correction to ADR 0011's consequences, though: it said a Partner Device's older history
could be fetched by querying the old REST timeseries endpoint directly, with an arbitrary
`from`/`to` range (up to 30 days). **The gRPC API has no equivalent** — `GetDeviceSnapshot`
only ever returns a fixed trailing 1-hour window, and there is no other RPC for wider
historical queries. This is a real capability loss, not just a rewording: once a Partner
Device reading is older than both the 1-hour window and this backend's own cache TTL, it is
gone — neither SMtrack's gRPC API nor this backend can recover it. No current consumer in
this codebase was found to rely on querying a wider window, but this is worth flagging to any
future feature that wants Partner Device trend/history data older than the cache TTL.
