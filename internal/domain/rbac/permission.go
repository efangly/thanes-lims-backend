package rbac

import (
	"fmt"
	"time"
)

// Module is the noun half of a Permission and of an Audit Entry's Resource:
// the business area being acted on. See CONTEXT.md#access-control.
type Module string

const (
	ModuleUser          Module = "user"
	ModuleSample        Module = "sample"
	ModuleTestResult    Module = "testresult"
	ModuleLocation      Module = "location"
	ModuleEquipment     Module = "equipment"
	ModuleInventory     Module = "inventory"
	ModulePurchaseOrder Module = "purchaseorder"
	ModuleDocument      Module = "document"
	ModuleEnvironment   Module = "environment"
	ModuleNotification  Module = "notification"
	ModuleAudit         Module = "audit"
)

// Action is the verb half of a Permission.
type Action string

const (
	ActionView    Action = "view"
	ActionCreate  Action = "create"
	ActionEdit    Action = "edit"
	ActionDelete  Action = "delete"
	ActionApprove Action = "approve"

	// ActionExport is granted only on ModuleAudit, distinct from
	// ActionView: a Role with audit:view can browse the JSON audit trail,
	// but only a Role with audit:export can pull the PDF compliance
	// export - Admin and QA get both, Lab Manager gets view only. See
	// CONTEXT.md#access-control and ADR 0002.
	ActionExport Action = "export"
)

// Permission is a grant of one Action on one Module to a Role. Permissions
// are never assigned to a User directly, only via their Role (see
// role_permissions). Truth lives in the permissions table - see
// migrations/000019_create_permissions_table.up.sql and ADR 0002.
type Permission struct {
	ID        int64
	Module    Module
	Action    Action
	CreatedAt time.Time
}

// Key is the compact "module:action" string embedded in the JWT access
// token's permissions claim - see ADR 0002.
func (p Permission) Key() string {
	return fmt.Sprintf("%s:%s", p.Module, p.Action)
}
