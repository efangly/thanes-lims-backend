package audit_test

import (
	"context"

	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
	"github.com/stretchr/testify/mock"
)

type mockAuditLogger struct{ mock.Mock }

func (m *mockAuditLogger) Log(ctx context.Context, entry domainaudit.AuditLog) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditLogger) List(ctx context.Context, filter portaudit.ListFilter) ([]domainaudit.AuditLog, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domainaudit.AuditLog), args.Error(1)
}
