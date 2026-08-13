package audit_test

import (
	"context"
	"testing"
	"time"

	applicationaudit "github.com/efangly/thanes-lims-backend/internal/application/audit"
	domainaudit "github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAuditLogsUseCase_PassesFilterThrough(t *testing.T) {
	logger := new(mockAuditLogger)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	filter := portaudit.ListFilter{From: &from}
	want := []domainaudit.AuditLog{{ID: 1, Method: "POST", Path: "/api/v1/samples"}}

	logger.On("List", mock.Anything, filter).Return(want, nil)

	uc := applicationaudit.NewListAuditLogsUseCase(logger)
	got, err := uc.Execute(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
