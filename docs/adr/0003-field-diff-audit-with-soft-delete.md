# Field-level diff audit entries backed by soft delete, not hard delete

`AuditLog` already has `Resource` and `Metadata` columns, but the global `Audit` middleware (`internal/adapters/http/middleware/audit.go`) never populates either — it only records method/path/status, so the audit trail can't currently answer "what changed." We're populating `Resource` with the Module name and `Metadata` with a Change Set (changed fields only, old→new) for a defined list of audited Modules (Sample, Sample CoC Step, Location, User, Test Result, Equipment/Calibration, Inventory Item, Purchase Order, Document). This requires reading a record's prior state before a write completes, which in turn requires that a delete not be a hard delete — otherwise the "after" state of a delete has nothing to diff against, and the deleted record's own Audit Trail becomes unreachable (a foreign key with nothing on the other end). We're adding a `deleted_at` column (Retired, per the domain glossary) to every audited entity instead.

Soft delete is the more expensive path day-to-day — every query on an audited entity must now exclude Retired rows, and uniqueness constraints (e.g. `users.email`) need to become partial indexes scoped to non-Retired rows so a Retired record doesn't permanently block reuse of a value. We accepted that cost because compliance-grade audit trails (this is a lab, samples and test results need to survive review/dispute after "deletion") can't have gaps where a record's history vanishes along with the record.

## Status

accepted

## Considered Options

- **Keep hard delete, log a full snapshot in the Audit Entry before deleting** — no schema change to the audited tables, but the record itself is gone: nothing to join back to from the Audit Entry, and any other table still holding a foreign key to it breaks.
- **Soft delete on every audited entity (chosen)** — the record and its full Audit Trail stay queryable and joinable indefinitely; pays for it with partial-unique-index bookkeeping and every read path needing a `deleted_at IS NULL` filter.
- **Full before/after snapshot per write instead of a Change Set** — simpler to compute (no diffing), but bloats `Metadata` with unchanged fields on every write and makes "what actually changed" a manual read-time diff instead of the stored fact.
