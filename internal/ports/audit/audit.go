package audit

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/audit"
)

// AuditLogger persists a single audit trail entry. Implementations must
// never block or fail the request they're auditing - errors are logged,
// not propagated to the caller's response.
type AuditLogger interface {
	Log(ctx context.Context, entry audit.AuditLog) error
	List(ctx context.Context, filter ListFilter) ([]audit.AuditLog, error)
}

// ListFilter narrows an export to a date range; a nil bound is open-ended.
type ListFilter struct {
	From *time.Time
	To   *time.Time
}
