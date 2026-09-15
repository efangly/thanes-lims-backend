# Partner API Guide — Third-Party Device Data Access

This is the integration guide for third-party applications reading device metadata and
telemetry (time series) from SMtrack. It is self-contained — you do not need access to the
internal frontend/admin API guide to use this API.

If you don't have an API key yet, contact your SMtrack admin. There is no self-service
sign-up; keys are issued manually.

## 1. Base URL

All endpoints in this guide are mounted under the `/log` prefix:

```
https://<your-smtrack-host>/log/partner/...
```

There is no `/v1` version segment yet. Breaking changes will be announced to partners ahead
of time rather than shipped silently.

## 2. Authentication

Every request must include your API key in a custom header:

```
X-API-Key: pk_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

- Do **not** send it via `Authorization` — that header is reserved for a different,
  internal auth scheme and is ignored by this API.
- Your key only grants access to a specific, pre-approved list of device serials (see
  below). It does not grant access to "all devices."
- Keys don't expire automatically. If a key needs to be rotated or revoked, contact your
  admin — a revoked key stops working immediately (`401`).
- **Keep your key secret.** Anyone with it can read data for every serial it's scoped to,
  within your rate limit.

## 3. What "serial" means here

The `serial` you pass in the URL identifies the **physical hardware unit** (the box), not a
building/room/location. If a unit is ever physically swapped out for a different one at the
same install location, its serial number changes — **you are responsible for tracking which
serial maps to which of your assets and updating your own records after a swap.** Ask your
admin to update your key's allowlist to include the new serial when that happens.

## 4. Rate limiting

Each API key is limited to **60 requests per minute**, across all endpoints combined. Going
over the limit returns:

```
HTTP 429 Too Many Requests
```

There is currently no per-key custom limit — all keys share the same fixed budget. If your
integration needs a higher limit, talk to your admin.

## 5. Response envelope

Every successful response is wrapped the same way as the rest of the SMtrack API:

```jsonc
{
  "success": true,
  "message": "Request Successfully",
  "data": { /* or [] for the time series endpoint */ },
  "meta": { "page": 1, "limit": 100, "total": 42, "totalPages": 1 }, // only on paginated endpoints
  "timestamp": "2026-09-14T07:27:47.943Z",
  "statusCode": 200
}
```

`data` holds the actual payload described per-endpoint below. `meta` is only present on the
time series endpoint (it's paginated); the metadata endpoint omits it.

## 6. Error envelope

Every error response (auth failure, validation failure, not found, etc.) has this shape:

```jsonc
{
  "success": false,
  "message": "Device SN-123 not found", // string, or an array of validation messages on 400
  "data": null
}
```

| Status | Meaning | Common cause |
|---|---|---|
| `400` | Bad request | Missing/invalid `from`/`to`, date range wider than 30 days, `from` after `to` |
| `401` | Unauthorized | Missing `X-API-Key` header, or the key is invalid/revoked |
| `404` | Not found | The `serial` doesn't exist, **or** it exists but isn't in your key's allowlist — these two cases are intentionally indistinguishable so a wrong/guessed serial can't be used to probe which devices exist |
| `429` | Too many requests | You exceeded 60 requests/minute for this key |

## 7. Endpoints

### `GET /log/partner/devices/{serial}` — device metadata

Returns a small, curated snapshot of the device's current state. Internal/location fields
(ward, hospital, physical position, etc.) are intentionally excluded — this endpoint never
returns anything about where the device is installed.

**Headers**

| Header | Required | Value |
|---|---|---|
| `X-API-Key` | yes | your API key |

**Response fields** (inside `data`)

| Field | Type | Description |
|---|---|---|
| `serial` | string | the serial you requested |
| `name` | string \| null | human-readable device name, if set |
| `status` | boolean | device enabled/active flag |
| `firmware` | string \| null | current firmware version reported by the hardware |
| `online` | boolean | whether the device is currently connected |

There is currently no "last seen" timestamp on this endpoint — if you need to know how
recent a device's data is, check the `sendTime` of its most recent time series point
instead (see below).

**Example**

```bash
curl -s "https://<host>/log/partner/devices/SN-00042" \
  -H "X-API-Key: $PARTNER_API_KEY"
```

```jsonc
{
  "success": true,
  "message": "Request Successfully",
  "data": {
    "serial": "SN-00042",
    "name": "Fridge — Ward 3",
    "status": true,
    "firmware": "1.4.2",
    "online": true
  },
  "timestamp": "2026-09-14T07:27:47.943Z",
  "statusCode": 200
}
```

---

### `GET /log/partner/devices/{serial}/timeseries` — telemetry history

Returns raw telemetry rows for the device, ordered newest first, paginated.

**Headers**

| Header | Required | Value |
|---|---|---|
| `X-API-Key` | yes | your API key |

**Query parameters**

| Param | Required | Type | Notes |
|---|---|---|---|
| `from` | **yes** | ISO 8601 date/datetime string | start of range (inclusive) |
| `to` | **yes** | ISO 8601 date/datetime string | end of range (inclusive) |
| `page` | no | integer, default `1` | 1-indexed |
| `limit` | no | integer, default `100`, max `500` | rows per page |

The `from`–`to` window **cannot exceed 30 days** per request — request `400` if it does.
For longer history, page through multiple 30-day windows.

**Response fields** (inside each item of `data[]`)

| Field | Type | Description |
|---|---|---|
| `serial` | string | matches the requested serial |
| `sendTime` | ISO datetime | when the reading was taken (device clock, normalized to UTC) |
| `temp` | number | raw temperature reading (°C) |
| `tempDisplay` | number | displayed/adjusted temperature (°C) |
| `humidity` | number | raw humidity reading (%) |
| `humidityDisplay` | number | displayed/adjusted humidity (%) |
| `tempInternal` | number \| null | internal sensor temperature, if applicable |
| `battery` | integer | battery level (%) |
| `plug` | boolean | mains power connected |
| `door1` / `door2` / `door3` | boolean | door sensor states |
| `internet` | boolean | device's own connectivity state at the time of the reading |
| `extMemory` | boolean | external storage present |
| `probe` | string | raw probe/channel identifier as reported by the device |
| `id`, `deviceId`, `probeId`, `createAt`, `updateAt` | — | internal bookkeeping fields — present today but not guaranteed; don't build logic around them |

**Example**

```bash
curl -s "https://<host>/log/partner/devices/SN-00042/timeseries?from=2026-08-01&to=2026-08-15&limit=2" \
  -H "X-API-Key: $PARTNER_API_KEY"
```

```jsonc
{
  "success": true,
  "message": "Request Successfully",
  "data": [
    {
      "serial": "SN-00042",
      "sendTime": "2026-08-15T00:00:00.000Z",
      "temp": 4.8,
      "tempDisplay": 4.8,
      "humidity": 52.1,
      "humidityDisplay": 52,
      "tempInternal": 23.1,
      "battery": 91,
      "plug": true,
      "door1": false,
      "door2": false,
      "door3": false,
      "internet": true,
      "extMemory": true,
      "probe": "1"
    }
  ],
  "meta": { "page": 1, "limit": 2, "total": 336, "totalPages": 168 },
  "timestamp": "2026-09-14T07:27:48.264Z",
  "statusCode": 200
}
```

## 8. Getting/rotating a key

There is no API for this — it's admin-only, out of band:

1. Tell your SMtrack admin which device serials you need access to.
2. They issue a key scoped to exactly those serials and hand you the plaintext value
   **once** — it is never shown again, so store it securely on your side immediately.
3. If you lose it or suspect it's leaked, ask your admin to revoke it and issue a new one.
