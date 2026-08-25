package user_test

import (
	"testing"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

// TestRole_Can and its Permission cases were removed along with the static
// Can() matrix (see ADR 0002 - permission truth now lives in the normalized
// roles/permissions/role_permissions tables, not a Go-code matrix).

func TestRole_Valid(t *testing.T) {
	assert.True(t, domainuser.RoleAdmin.Valid())
	assert.True(t, domainuser.RoleLabManager.Valid())
	assert.False(t, domainuser.Role("bogus").Valid())
}

func TestRole_DisplayName(t *testing.T) {
	assert.Equal(t, "Admin", domainuser.RoleAdmin.DisplayName())
	assert.Equal(t, "Lab Manager", domainuser.RoleLabManager.DisplayName())
	assert.Equal(t, "QA", domainuser.RoleQA.DisplayName())
	assert.Equal(t, "Scientist", domainuser.RoleScientist.DisplayName())
	assert.Equal(t, "General", domainuser.RoleGeneral.DisplayName())
}
