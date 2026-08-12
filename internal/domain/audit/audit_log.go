package audit

import "time"

// AuditLog is an append-only record of a write operation performed against
// the system, captured centrally by HTTP middleware for every POST/PATCH/DELETE.
type AuditLog struct {
	ID         int64
	ActorID    *int64
	ActorRole  string
	Method     string
	Path       string
	Resource   string
	ResourceID string
	StatusCode int
	Metadata   map[string]any
	CreatedAt  time.Time
}
