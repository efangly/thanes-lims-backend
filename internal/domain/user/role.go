package user

// Role is the system-wide RBAC role. Permissions are global-per-role (not
// module-scoped) - every module checks the same matrix below.
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleQA        Role = "qa"
	RoleScientist Role = "scientist"
	RoleGeneral   Role = "general"
)

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleQA, RoleScientist, RoleGeneral:
		return true
	default:
		return false
	}
}

// Permission mirrors the frontend's RBAC demo matrix (ดู/แก้ไข/ลบ/อนุมัติ).
type Permission string

const (
	PermView    Permission = "view"
	PermEdit    Permission = "edit"
	PermDelete  Permission = "delete"
	PermApprove Permission = "approve"
)

// permissionMatrix is the static global-per-role permission table, ported
// directly from the frontend's manage-access.tsx demo matrix.
var permissionMatrix = map[Role]map[Permission]bool{
	RoleAdmin:     {PermView: true, PermEdit: true, PermDelete: true, PermApprove: true},
	RoleQA:        {PermView: true, PermEdit: true, PermDelete: false, PermApprove: true},
	RoleScientist: {PermView: true, PermEdit: true, PermDelete: false, PermApprove: false},
	RoleGeneral:   {PermView: true, PermEdit: false, PermDelete: false, PermApprove: false},
}

func (r Role) Can(perm Permission) bool {
	perms, ok := permissionMatrix[r]
	if !ok {
		return false
	}
	return perms[perm]
}
