# Partner Device readings are proxied, not persisted

Partner Device telemetry (temp/humidity from SMtrack's Partner API, `docs/partner-api-guide.md`) is polled every 30s by our own backend, cached briefly (30-60s TTL, shared per Serial), and evaluated against the linked Location's Gauge thresholds — but the raw readings themselves are never written to `sensor_readings`. Only the derived Env Alert persists, same as for any other Location.

Considered persisting every polled reading as a Sensor Reading (consistent with how other Locations' history works), but rejected it: SMtrack already stores and owns the full history behind its own paginated `timeseries` endpoint (max 30-day window per request), so persisting would mean carrying a second, redundant copy of data we don't control the retention or correctness of, plus backfill/catch-up logic for gaps. SMtrack stays the system of record for Partner Device raw history; this backend is a live proxy/aggregator over it, not a mirror.

Consequence: a Partner Device Location's trend/history view cannot rely on `sensor_readings` for anything older than the cache TTL — historical queries for a Partner Device must go back to the Partner API's `timeseries` endpoint directly (subject to its 30-day window and 60 req/min rate limit), unlike other Locations whose full history lives in Postgres.
