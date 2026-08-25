package rbac

// RolePermission is a single grant: the Role identified by RoleID holds the
// Permission identified by PermissionID. Truth lives in the
// role_permissions join table - see
// migrations/000020_create_role_permissions_table.up.sql and ADR 0002.
type RolePermission struct {
	RoleID       int64
	PermissionID int64
}
