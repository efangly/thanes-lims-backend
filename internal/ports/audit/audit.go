package audit

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/audit"
)

// AuditLogger persists a single audit trail entry. Implementations must
// never block or fail the request they're auditing - errors are logged,
// not propagated to the caller's response.
type AuditLogger interface {
	Log(ctx context.Context, entry audit.AuditLog) error
}
