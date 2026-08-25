package audit

import (
	"context"

	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
)

// ListAuditLogsPage is the result of a paginated audit trail query: the
// page of entries plus the total count matching the filter (ignoring
// Limit/Offset), used to compute pagination meta in the HTTP response.
type ListAuditLogsPage struct {
	Entries []domainaudit.AuditLog
	Total   int64
}

// ListAuditLogsPageUseCase backs the JSON GET /audit/logs endpoint. It's a
// sibling to ListAuditLogsUseCase (which backs the PDF /audit/export and
// always pulls the full matching set) because this one also needs a total
// count for pagination.
type ListAuditLogsPageUseCase struct {
	logger portaudit.AuditLogger
}

func NewListAuditLogsPageUseCase(logger portaudit.AuditLogger) *ListAuditLogsPageUseCase {
	return &ListAuditLogsPageUseCase{logger: logger}
}

func (uc *ListAuditLogsPageUseCase) Execute(ctx context.Context, filter portaudit.ListFilter) (ListAuditLogsPage, error) {
	entries, err := uc.logger.List(ctx, filter)
	if err != nil {
		return ListAuditLogsPage{}, err
	}
	total, err := uc.logger.Count(ctx, filter)
	if err != nil {
		return ListAuditLogsPage{}, err
	}
	return ListAuditLogsPage{Entries: entries, Total: total}, nil
}
