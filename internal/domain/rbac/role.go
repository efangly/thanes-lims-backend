package rbac

import "time"

// Role is a named set of Permissions assigned to a User (exactly one Role
// per User). Truth lives in the roles table - see
// migrations/000018_create_roles_table.up.sql and ADR 0002.
type Role struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
