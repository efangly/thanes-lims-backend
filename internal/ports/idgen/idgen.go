package idgen

import "context"

// SequenceGenerator issues the next atomic counter value for a named scope
// (e.g. "sample", "equipment"). Scopes that are Buddhist-year-scoped (Sample,
// PurchaseOrder) pass a non-nil year so counters reset per year; scopes that
// aren't pass nil. Formatting the raw integer into a human-readable ID
// ("SMP-2569-00001") is the caller's (application layer's) responsibility -
// this port only guarantees a unique, monotonically increasing number.
type SequenceGenerator interface {
	Next(ctx context.Context, scope string, year *int) (int64, error)
}
