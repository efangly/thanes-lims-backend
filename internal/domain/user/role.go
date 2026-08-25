package user

// Role is the system-wide RBAC role assigned to a User (exactly one Role
// per User). Permission truth no longer lives here as a static matrix - it
// lives in the normalized roles/permissions/role_permissions tables (see
// internal/domain/rbac and ADR 0002). Role is kept as this fixed enum
// because the JWT "role" claim still keys off it, and it's the type used to
// persist/read users.role_id.
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleLabManager Role = "lab_manager"
	RoleQA         Role = "qa"
	RoleScientist  Role = "scientist"
	RoleGeneral    Role = "general"
)

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleLabManager, RoleQA, RoleScientist, RoleGeneral:
		return true
	default:
		return false
	}
}

// DisplayName is the roles.name row this Role corresponds to in the
// database (see migrations/000018_create_roles_table.up.sql). Used to
// translate between this Go enum and the normalized RBAC tables - e.g.
// resolving a User's Permission set at login, or persisting/reading
// users.role_id.
func (r Role) DisplayName() string {
	switch r {
	case RoleAdmin:
		return "Admin"
	case RoleLabManager:
		return "Lab Manager"
	case RoleQA:
		return "QA"
	case RoleScientist:
		return "Scientist"
	case RoleGeneral:
		return "General"
	default:
		return string(r)
	}
}
