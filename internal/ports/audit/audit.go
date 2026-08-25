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
	Count(ctx context.Context, filter ListFilter) (int64, error)
}

// ListFilter narrows a listing/export by date range and, for the JSON
// browse endpoint, by actor/resource/method. A nil/empty field is
// unconstrained. Limit/Offset of 0 mean "no pagination" (used by the PDF
// export, which always pulls the full matching set).
type ListFilter struct {
	From     *time.Time
	To       *time.Time
	ActorID  *int64
	Resource string
	Method   string
	Limit    int
	Offset   int
}
