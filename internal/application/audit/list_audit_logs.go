package audit

import (
	"context"

	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
)

type ListAuditLogsUseCase struct {
	logger portaudit.AuditLogger
}

func NewListAuditLogsUseCase(logger portaudit.AuditLogger) *ListAuditLogsUseCase {
	return &ListAuditLogsUseCase{logger: logger}
}

func (uc *ListAuditLogsUseCase) Execute(ctx context.Context, filter portaudit.ListFilter) ([]domainaudit.AuditLog, error) {
	return uc.logger.List(ctx, filter)
}
