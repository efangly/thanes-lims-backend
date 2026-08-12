package audit

import (
	"context"

	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
)

// LogActionUseCase wraps the AuditLogger port so the audit middleware
// depends on the application layer rather than reaching into adapters
// directly, keeping the module consistent and testable with the rest of
// the codebase even though the logic here is a thin pass-through.
type LogActionUseCase struct {
	logger portaudit.AuditLogger
}

func NewLogActionUseCase(logger portaudit.AuditLogger) *LogActionUseCase {
	return &LogActionUseCase{logger: logger}
}

func (uc *LogActionUseCase) Execute(ctx context.Context, entry domainaudit.AuditLog) error {
	return uc.logger.Log(ctx, entry)
}
