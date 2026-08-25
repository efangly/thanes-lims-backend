package rbac

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
)

// Repository is the read path RBAC needs at login/refresh: resolving a
// Role's full Permission set for embedding into the JWT access token (see
// ADR 0002). Role/Permission/RolePermission management is not part of this
// phase - only this one read path.
type Repository interface {
	FindPermissionsByRoleName(ctx context.Context, roleName string) ([]rbac.Permission, error)
}
