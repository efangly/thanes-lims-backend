# Partner API Guide — SMtrack Third-Party Device Data (gRPC)

This is the integration guide for this backend's outbound connection to SMtrack's Partner
API, which reads device metadata and telemetry for admin-mapped Partner Devices (see
`CONTEXT.md#environment`, ADR 0011, ADR 0012).

**This surface is gRPC, not REST** — SMtrack retired the REST version of this API; there is
no fallback. The full contract lives in
[`proto/partner/partner.proto`](../proto/partner/partner.proto) in this repo — that file (plus
`option go_package`, added here for codegen) is the source of truth; the tables below are for
reference only.

## 1. Connection

```
addr   PARTNER_GRPC_ADDR (bare host:port, e.g. siamatic.thddns.net:50051) — NOT a URL,
       no http(s):// scheme
TLS    plaintext for now — SMtrack has not yet configured TLS on this endpoint; treat this
       as a known, accepted gap (see ADR 0012), not a deliberate choice
```

`internal/adapters/partnergrpc.New` strips an accidentally-pasted `http://`/`https://`
prefix defensively, but the correct `.env` value has no scheme.

## 2. Authentication

Every call carries the API key as gRPC **metadata** (not a message field, not an HTTP
header):

```
x-api-key: pk_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

The key (`PARTNER_API_KEY`) only grants access to a pre-approved list of **wards**
(departments/units) — it is not scoped by individual device serial. Any device that belongs
to (or is later moved into) an allowed ward becomes visible automatically.

## 3. Rate limiting

60 requests/minute per key, across both RPCs combined. Over the limit returns
`RESOURCE_EXHAUSTED` (8).

## 4. Error codes

| gRPC status | Code | Meaning | Treated as (see ADR 0011) |
|---|---|---|---|
| `UNAUTHENTICATED` | 16 | Missing, invalid, or revoked `x-api-key` | permanent |
| `NOT_FOUND` | 5 | Serial/ward doesn't exist, or isn't in this key's allowed wards | permanent |
| `INVALID_ARGUMENT` | 3 | Empty `serial`/`ward`, or a bad `page`/`limit` | permanent |
| `RESOURCE_EXHAUSTED` | 8 | Over 60 requests/minute | retryable |
| `UNAVAILABLE` | 14 | Network/connectivity problem | retryable |
| `DEADLINE_EXCEEDED` | 4 | Call timed out | retryable |

"Permanent" vs "retryable" here is this backend's own classification
(`partnergrpc.RPCError.Retryable()`), driving the poller's stale-cache-fallback decision —
not something SMtrack's API documents itself.

## 5. RPCs

### `GetDeviceSnapshot(serial)` — metadata + trailing 1-hour telemetry

Returns device metadata plus the last hour of telemetry, always — there is no way to
request a wider or different window; this RPC is fixed-window, not paginated history. This
backend only consumes `timeseries[0]` (the newest point) as the "latest reading"; the rest
of the hour is not currently surfaced anywhere in this backend.

| Field | Type | Notes |
|---|---|---|
| `device.serial/name/status/firmware/online` | | curated metadata, no location/ward fields |
| `timeseries` | repeated `TelemetryPoint` | last 1h, newest first — empty if no readings |

### `ListDevicesByWard(ward, page, limit)` — every device in a ward + its latest reading

Returns every device in `ward` that the key is scoped to (paginated), each with metadata
plus its single most recent reading (absent, not zero, if the device has no telemetry yet).
This backend uses this RPC for the admin-facing `GET /partner-devices/discover` endpoint —
browsing SMtrack's inventory to find a device's `serial` before creating a Partner Device
(Serial↔Location) mapping here. It is read-through only; results are never persisted (ADR
0011).

## 6. What changed from the old REST API

- Transport: REST (`X-API-Key` header, `{success,message,data}` envelope, HTTP status codes)
  → gRPC (`x-api-key` metadata, gRPC status codes). See ADR 0012.
- `GetDeviceSnapshot` used to require two REST calls (metadata, then a timeseries call for
  the latest reading); the gRPC version returns both in one call.
- **Capability lost**: the old REST timeseries endpoint supported an arbitrary `from`/`to`
  date range (up to 30 days). The gRPC API has no equivalent — `GetDeviceSnapshot` only ever
  returns a fixed trailing 1-hour window, and there is no other RPC for historical queries.
  ADR 0011's original "go back to the Partner API directly for older history" fallback no
  longer has anywhere to go for this backend's own trend/history views once the 1-hour
  window and the local cache TTL have both passed.
- New capability: `ListDevicesByWard`, with no REST precedent.
