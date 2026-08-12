package user_test

import (
	"testing"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestRole_Can(t *testing.T) {
	cases := []struct {
		role domainuser.Role
		perm domainuser.Permission
		want bool
	}{
		{domainuser.RoleAdmin, domainuser.PermDelete, true},
		{domainuser.RoleAdmin, domainuser.PermApprove, true},
		{domainuser.RoleQA, domainuser.PermApprove, true},
		{domainuser.RoleQA, domainuser.PermDelete, false},
		{domainuser.RoleScientist, domainuser.PermEdit, true},
		{domainuser.RoleScientist, domainuser.PermApprove, false},
		{domainuser.RoleGeneral, domainuser.PermView, true},
		{domainuser.RoleGeneral, domainuser.PermEdit, false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.role.Can(tc.perm), "role=%s perm=%s", tc.role, tc.perm)
	}
}

func TestRole_Valid(t *testing.T) {
	assert.True(t, domainuser.RoleAdmin.Valid())
	assert.False(t, domainuser.Role("bogus").Valid())
}
